package apps

import (
	"cloudiac/common"
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/libs/page"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
	"cloudiac/portal/services/logstorage"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// CreatePolicy 创建策略
func CreatePolicy(c *ctx.ServiceContext, form *forms.CreatePolicyForm) (*models.Policy, e.Error) {
	c.AddLogField("action", fmt.Sprintf("create policy %s", form.Name))

	ruleName, resourceType, policyType, err := parseRegoHeader(form.Rego)
	if err != nil {
		return nil, e.New(err.Code(), err, http.StatusBadRequest)
	}

	p := models.Policy{
		Name:          form.Name,
		CreatorId:     c.UserId,
		FixSuggestion: form.FixSuggestion,
		Severity:      form.Severity,
		Rego:          form.Rego,
		RuleName:      ruleName,
		ResourceType:  resourceType,
		PolicyType:    policyType,
	}
	refId, err := services.GetPolicyReferenceId(c.DB(), &p)
	if err != nil {
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	}
	p.ReferenceId = refId

	policy, err := services.CreatePolicy(c.DB(), &p)
	if err != nil && err.Code() == e.PolicyAlreadyExist {
		return nil, e.New(err.Code(), err, http.StatusBadRequest)
	} else if err != nil {
		c.Logger().Errorf("error creating policy, err %s", err)
		return nil, e.AutoNew(err, e.DBError)
	}

	return policy, nil
}

//parseRegoHeader 解析 rego 脚本获取入口，云商类型和资源类型
func parseRegoHeader(rego string) (ruleName string, policyType string, resType string, err e.Error) {
	regex := regexp.MustCompile("(?m)^##ruleName (.*)$")
	match := regex.FindStringSubmatch(rego)
	if len(match) == 2 {
		ruleName = strings.TrimSpace(match[1])
	} else {
		return "", "", "", e.New(e.PolicyRegoMissingComment, fmt.Errorf("missing ##ruleName comment"))
	}

	regex = regexp.MustCompile("(?m)^##policyType (.*)$")
	match = regex.FindStringSubmatch(rego)
	if len(match) == 2 {
		policyType = strings.TrimSpace(match[1])
	} else {
		return "", "", "", e.New(e.PolicyRegoMissingComment, fmt.Errorf("missing ##policyType comment"))
	}

	regex = regexp.MustCompile("(?m)^##resourceType (.*)$")
	match = regex.FindStringSubmatch(rego)
	if len(match) == 2 {
		resType = strings.TrimSpace(match[1])
	} else {
		return "", "", "", e.New(e.PolicyRegoMissingComment, fmt.Errorf("missing ##resourceType comment"))
	}
	return
}

// ScanTemplate 扫描云模板策略
func ScanTemplate(c *ctx.ServiceContext, form *forms.ScanTemplateForm, envId models.Id) (*models.ScanTask, e.Error) {
	c.AddLogField("action", fmt.Sprintf("scan template %s", form.Id))
	if c.OrgId == "" {
		return nil, e.New(e.BadRequest, http.StatusBadRequest)
	}

	tx := c.Tx()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	var (
		env *models.Env
		err e.Error
	)

	orgQuery := services.QueryWithOrgId(tx, c.OrgId)

	// 环境检查
	if envId != "" {
		env, err = services.GetEnvById(orgQuery, envId)
		if err != nil && err.Code() == e.EnvNotExists {
			return nil, e.New(err.Code(), err, http.StatusBadRequest)
		} else if err != nil {
			c.Logger().Errorf("error get environment, err %s", err)
			return nil, e.New(e.DBError, err, http.StatusInternalServerError)
		}
	}

	// 模板检查
	tpl, err := services.GetTemplateById(orgQuery, form.Id)
	if err != nil && err.Code() == e.TemplateNotExists {
		return nil, e.New(err.Code(), err, http.StatusBadRequest)
	} else if err != nil {
		c.Logger().Errorf("error get template, err %s", err)
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	}

	// 创建任务
	runnerId, err := services.GetDefaultRunnerId()
	if err != nil {
		return nil, e.New(err.Code(), err, http.StatusInternalServerError)
	}
	taskType := models.TaskTypeScan
	if form.Parse {
		taskType = models.TaskTypeParse
	}
	task, err := services.CreateScanTask(tx, tpl, env, models.ScanTask{
		Name:      models.ScanTask{}.GetTaskNameByType(taskType),
		CreatorId: c.UserId,
		TplId:     tpl.Id,
		EnvId:     envId,
		Revision:  tpl.RepoRevision,
		BaseTask: models.BaseTask{
			Type:        taskType,
			Flow:        models.TaskFlow{},
			StepTimeout: common.TaskStepTimeoutDuration,
			RunnerId:    runnerId,
		},
	})
	if err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("error creating scan task, err %s", err)
		return nil, e.New(err.Code(), err, http.StatusInternalServerError)
	}

	if _, err := tx.Save(task); err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("error save env, err %s", err)
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("error commit env, err %s", err)
		return nil, e.New(e.DBError, err)
	}

	return task, nil
}

// ScanEnvironment 扫描环境策略
func ScanEnvironment(c *ctx.ServiceContext, form *forms.ScanEnvironmentForm) (*models.ScanTask, e.Error) {
	c.AddLogField("action", fmt.Sprintf("scan environment %s", form.Id))
	if c.OrgId == "" {
		return nil, e.New(e.BadRequest, http.StatusBadRequest)
	}

	env, err := services.GetEnvById(c.DB(), form.Id)
	if err != nil {
		return nil, e.New(err.Code(), err, http.StatusBadRequest)
	}

	f := forms.ScanTemplateForm{
		Id: env.TplId,
	}

	return ScanTemplate(c, &f, env.Id)
}

type PolicyResp struct {
	models.Policy
	GroupName  string `json:"groupName"`
	Creator    string `json:"creator"`
	Passed     int    `json:"passed"`
	Violated   int    `json:"violated"`
	Suppressed int    `json:"suppressed"`
	Failed     int    `json:"failed"`
}

// SearchPolicy 查询策略组列表
func SearchPolicy(c *ctx.ServiceContext, form *forms.SearchPolicyForm) (interface{}, e.Error) {
	query := services.SearchPolicy(c.DB(), form)
	policyResps := make([]PolicyResp, 0)
	p := page.New(form.CurrentPage(), form.PageSize(), query)
	if err := p.Scan(&policyResps); err != nil {
		return nil, e.New(e.DBError, err)
	}

	// 扫描结果统计信息
	var policyIds []models.Id
	for idx := range policyResps {
		policyIds = append(policyIds, policyResps[idx].Id)
	}
	if summaries, err := services.PolicySummary(c.DB(), policyIds, consts.ScopePolicy); err != nil {
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	} else if summaries != nil && len(summaries) > 0 {
		sumMap := make(map[string]*services.PolicyScanSummary, len(policyIds))
		for idx, summary := range summaries {
			sumMap[string(summary.Id)+summary.Status] = summaries[idx]
		}
		for idx, policyResp := range policyResps {
			if summary, ok := sumMap[string(policyResp.Id)+"passed"]; ok {
				policyResps[idx].Passed = summary.Count
			}
			if summary, ok := sumMap[string(policyResp.Id)+"violated"]; ok {
				policyResps[idx].Violated = summary.Count
			}
			if summary, ok := sumMap[string(policyResp.Id)+"failed"]; ok {
				policyResps[idx].Failed = summary.Count
			}
			if summary, ok := sumMap[string(policyResp.Id)+"suppressed"]; ok {
				policyResps[idx].Suppressed = summary.Count
			}
		}
	}

	return policyResps, nil
}

// UpdatePolicy 修改策略组
func UpdatePolicy(c *ctx.ServiceContext, form *forms.UpdatePolicyForm) (interface{}, e.Error) {
	attr := models.Attrs{}
	if form.HasKey("name") {
		attr["name"] = form.Name
	}

	if form.HasKey("fixSuggestion") {
		attr["fixSuggestion"] = form.FixSuggestion
	}

	if form.HasKey("severity") {
		attr["severity"] = form.Severity
	}

	if form.HasKey("rego") {
		attr["rego"] = form.Rego
	}

	if form.HasKey("status") {
		attr["status"] = form.Status
	}

	pg := models.Policy{}
	pg.Id = form.Id
	if _, err := services.UpdatePolicy(c.DB(), &pg, attr); err != nil {
		return nil, err
	}
	return nil, nil
}

// DeletePolicy 删除策略组
func DeletePolicy(c *ctx.ServiceContext, form *forms.DeletePolicyForm) (interface{}, e.Error) {
	return services.DeletePolicy(c.DB(), form.Id)

}

// DetailPolicy 查询策略组详情
func DetailPolicy(c *ctx.ServiceContext, form *forms.DetailPolicyForm) (interface{}, e.Error) {
	return services.DetailPolicy(c.DB(), form.Id)
}

// CreatePolicySuppress 策略屏蔽
func CreatePolicySuppress(c *ctx.ServiceContext, form *forms.CreatePolicyShieldForm) (interface{}, e.Error) {
	return services.CreatePolicySuppress()
}

func SearchPolicySuppress(c *ctx.ServiceContext, form *forms.SearchPolicySuppressForm) (interface{}, e.Error) {
	return services.SearchPolicySuppress()
}

func DeletePolicySuppress(c *ctx.ServiceContext, form *forms.DeletePolicySuppressForm) (interface{}, e.Error) {
	return services.DeletePolicySuppress()
}

type RespPolicyTpl struct {
	TplName         string    `json:"tplName" form:"tplName" `
	TplId           models.Id `json:"tplId" form:"tplId" `
	RepoAddr        string    `json:"repoAddr" form:"repoAddr" `
	PolicyGroupName string    `json:"policyGroupName" form:"policyGroupName" `
	PolicyGroupId   models.Id `json:"policyGroupId" form:"policyGroupId" `
	Enabled         bool      `json:"enabled" example:"true"`          //是否启用
	GroupStatus     string    `json:"groupStatus" form:"groupStatus" ` //状态 todo 不确认字段
}

func SearchPolicyTpl(c *ctx.ServiceContext, form *forms.SearchPolicyTplForm) (interface{}, e.Error) {
	return services.SearchPolicyTpl()
}

func UpdatePolicyTpl(c *ctx.ServiceContext, form *forms.UpdatePolicyTplForm) (interface{}, e.Error) {
	return services.UpdatePolicyTpl()
}

func DetailPolicyTpl(c *ctx.ServiceContext, form *forms.DetailPolicyTplForm) (interface{}, e.Error) {
	return services.DetailPolicyTpl()
}

type RespPolicyEnv struct {
	TplName         string    `json:"tplName" form:"tplName" `
	TplId           models.Id `json:"tplId" form:"tplId" `
	EnvName         string    `json:"envName" form:"envName" `
	EnvId           models.Id `json:"envId" form:"envId" `
	RepoAddr        string    `json:"repoAddr" form:"repoAddr" `
	PolicyGroupName string    `json:"policyGroupName" form:"policyGroupName" `
	PolicyGroupId   models.Id `json:"policyGroupId" form:"policyGroupId" `
	Enabled         bool      `json:"enabled" example:"true"`          //是否启用
	GroupStatus     string    `json:"groupStatus" form:"groupStatus" ` //状态 todo 不确认字段
}

func SearchPolicyEnv(c *ctx.ServiceContext, form *forms.SearchPolicyEnvForm) (interface{}, e.Error) {
	return services.SearchPolicyEnv()
}

func UpdatePolicyEnv(c *ctx.ServiceContext, form *forms.UpdatePolicyEnvForm) (interface{}, e.Error) {
	return services.UpdatePolicyEnv()
}

func DetailPolicyEnv(c *ctx.ServiceContext, form *forms.DetailPolicyEnvForm) (interface{}, e.Error) {
	return services.DetailPolicyEnv()
}

func PolicyError(c *ctx.ServiceContext, form *forms.PolicyErrorForm) (interface{}, e.Error) {
	return services.PolicyError()
}

func PolicyReference(c *ctx.ServiceContext, form *forms.PolicyReferenceForm) (interface{}, e.Error) {
	return services.PolicyReference()
}

func PolicyRepo(c *ctx.ServiceContext, form *forms.PolicyRepoForm) (interface{}, e.Error) {
	return services.PolicyRepo()
}

type ParseResp struct {
	Template *services.TfParse `json:"template"`
}

// ParseTemplate 解析云模板/环境源码
func ParseTemplate(c *ctx.ServiceContext, form *forms.PolicyParseForm) (interface{}, e.Error) {
	c.AddLogField("action", fmt.Sprintf("parse template %s env %s", form.TemplateId, form.EnvId))

	tplId := form.TemplateId
	envId := models.Id("")
	if form.HasKey("envId") {
		env, err := services.GetEnvById(c.DB(), form.EnvId)
		if err != nil {
			return nil, e.New(err.Code(), err, http.StatusBadRequest)
		}
		tplId = env.TplId
		envId = env.Id
	}

	f := forms.ScanTemplateForm{
		Id:    tplId,
		Parse: true,
	}
	scanTask, err := ScanTemplate(c, &f, envId)
	if err != nil {
		return nil, err
	}

	ticker := time.NewTicker(time.Second)
	timeout := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	defer timeout.Stop()

	// 等待任务执行完成
	for {
		scanTask, err = services.GetScanTaskById(c.DB(), scanTask.Id)
		if scanTask.IsExitedStatus(scanTask.Status) {
			break
		}

		select {
		case <-ticker.C:
			continue
		case <-timeout.C:
			return nil, e.New(e.PolicyErrorParseTemplate, fmt.Errorf("parse tempalte timeout"), http.StatusInternalServerError)
		}
	}

	if scanTask.Status == common.TaskComplete {
		content, er := logstorage.Get().Read(scanTask.TfParseJsonPath())
		if er != nil {
			return nil, e.New(e.PolicyErrorParseTemplate, fmt.Errorf("parse tempalte error: %v", err), http.StatusInternalServerError)
		}
		js, err := services.UnmarshalTfParseJson(content)
		if err != nil {
			return nil, e.New(e.PolicyErrorParseTemplate, fmt.Errorf("parse tempalte error: %v", err), http.StatusInternalServerError)
		}
		return ParseResp{
			Template: js,
		}, nil
	}
	return nil, e.New(e.PolicyErrorParseTemplate, fmt.Errorf("execute parse tempalte error: %v", err), http.StatusInternalServerError)
}

type ScanResultResp struct {
	ScanTime    *models.Time   `json:"scanTime"`
	ScanResults []PolicyResult `json:"scanResults"`
}

type PolicyResult struct {
	models.PolicyResult
	PolicyName      string `json:"policyName"`
	PolicyGroupName string `json:"policyGroupName"`
	FixSuggestion   string `json:"fixSuggestion"`
}

func PolicyScanResult(c *ctx.ServiceContext, form *forms.PolicyScanResultForm) (*ScanResultResp, e.Error) {
	c.AddLogField("action", fmt.Sprintf("scan result %s %s", form.Scope, form.Id))
	var (
		envId models.Id
		tplId models.Id
	)
	if form.Scope == consts.ScopeEnv {
		envId = form.Id
	} else {
		tplId = form.Id
	}
	scanTask, err := services.GetLastScanTask(c.DB(), envId, tplId)
	if err != nil {
		if err.Code() != e.TaskNotExists {
			return nil, nil
		} else {
			return nil, err
		}
	}

	query := services.QueryPolicyResult(c.DB(), scanTask.Id)
	if form.SortField() == "" {
		query = query.Order("policy_group_name, policy_name")
	} else {
		query = form.Order(query)
	}
	results := make([]PolicyResult, 0)
	p := page.New(form.CurrentPage(), form.PageSize(), query)
	if err := p.Scan(&results); err != nil {
		return nil, e.New(e.DBError, err)
	}

	return &ScanResultResp{
		ScanTime:    scanTask.StartAt,
		ScanResults: results,
	}, nil
}

type ScanReport struct {
	Summary       Summary            `json:"summary"`
	ScannedByDays ScanReportDay      `json:"scannedByDays"`
	PassedByDays  ScanReportDay      `json:"passedByDays"`
	LastScanned   []LastScanTaskResp `json:"lastScanned"`
}

type Summary struct {
	Passed   int `json:"passed"` // 百分比，按 100 = 100%
	Violated int `json:"violated"`
	Suppress int `json:"suppress"`
	Failed   int `json:"failed"`
}

type ScanReportDay struct {
	Day   string `json:"date"`
	Count int    `json:"count"`
}

type LastScanTaskResp struct {
	models.ScanTask
	Scope       string `json:"scope"`       // 目标类型：环境/模板
	Name        string `json:"name"`        //检查目标
	OrgName     string `json:"orgName"`     // 组织名称
	ProjectName string `json:"projectName"` // 项目
	Creator     string `json:"creator"`     // 创建者
	Summary
}

type PolicyScanReportResp struct {
	PassedRate Polyline `json:"passedRate"` // 检测通过率
}

type Polyline struct {
	Column []string      `json:"col" example:"[\"08/20\",\"08-21\"]"`
	Value  []interface{} `json:"value" example:"[0.95,0.96]" swaggertype:"string"`
}

func PolicyScanReport(c *ctx.ServiceContext, form *forms.PolicyScanReportForm) (*PolicyScanReportResp, e.Error) {
	// data: {
	//  x轴坐标： [ 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun' ],
	//  数据：[ 12, 22, 33, 55, 77, 88, 99 ]
	//}

	return nil, nil
}

type PolicyGroupScanReportResp struct {
	Total            PieChar  `json:"total"`            // 检测结果比例
	TaskScanCount    Polyline `json:"scanCount"`        // 检测源执行次数
	PolicyScanCount  Polyline `json:"policyScanCount"`  // 策略运行趋势
	PolicyPassedRate Polyline `json:"policyPassedRate"` // 检测通过率趋势
}

type PieChar []PieSector

type PieSector struct {
	Name  string      `json:"name" example:"08-20"`
	Value interface{} `json:"value" example:"10" swaggertype:"string"`
}

func PolicyGroupScanReport(c *ctx.ServiceContext, form *forms.PolicyScanReportForm) (*PolicyScanReportResp, e.Error) {
	// data: [
	//        { value: 335, name: '通过' },
	//        { value: 310, name: '未通过' },
	//        { value: 234, name: '屏蔽' }
	// ]

	return nil, nil
}

func PolicyLastTasks(c *ctx.ServiceContext, form *forms.PolicyLastTasksForm) ([]*LastScanTaskResp, e.Error) {

	return nil, nil
}

type PolicyTestResp struct {
	Data  interface{} `json:"data" example:"{\n\"accurics\":{\n\"instanceWithNoVpc\":[\n{\n\"Id\":\"alicloud_instance.instance\"\n}\n]\n}\n}"` // 脚本测试输出，json文本
	Error string      `json:"error" example:"1 error occurred: policy.rego:4: rego_parse_error: refs cannot be used for rule\n"`               // 脚本执行错误内容
}

func PolicyTest(c *ctx.ServiceContext, form *forms.PolicyTestForm) (*PolicyTestResp, e.Error) {
	c.AddLogField("action", fmt.Sprintf("test template"))

	if _, _, _, err := parseRegoHeader(form.Rego); err != nil {
		return &PolicyTestResp{
			Data:  "",
			Error: fmt.Sprintf("1 error occurred: %s", err.Error()),
		}, nil
	}
	if err := json.Unmarshal([]byte(form.Input), nil); err != nil {
		return &PolicyTestResp{
			Data:  "",
			Error: fmt.Sprintf("invalid input"),
		}, nil
	}

	tmpDir, err := os.MkdirTemp("", "*")
	if err != nil {
		return nil, e.New(e.InternalError, errors.Wrapf(err, "create tmp dir"), http.StatusInternalServerError)
	}
	defer os.RemoveAll(tmpDir)

	regoPath := filepath.Join(tmpDir, "policy.rego")
	inputPath := filepath.Join(tmpDir, "input.json")

	if err := os.WriteFile(regoPath, []byte(form.Rego), 0644); err != nil {
		return nil, e.New(e.InternalError, err, http.StatusInternalServerError)
	}
	if err := os.WriteFile(inputPath, []byte(form.Input), 0644); err != nil {
		return nil, e.New(e.InternalError, err, http.StatusInternalServerError)
	}

	cmdline := fmt.Sprintf("cd %s && opa eval -f pretty --data %s --input %s data", tmpDir, "policy.rego", "input.json")
	cmd := exec.Command("sh", "-c", cmdline)
	if output, err := cmd.CombinedOutput(); err != nil {
		return &PolicyTestResp{
			Data:  "",
			Error: string(output),
		}, nil
	} else {
		return &PolicyTestResp{
			Data:  output,
			Error: "",
		}, nil
	}
}
