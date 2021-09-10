// Copyright 2021 CloudJ Company Limited. All rights reserved.

package apps

import (
	"cloudiac/common"
	"cloudiac/policy"
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
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
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
		Tags:          form.Tags,
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

	// 环境检查
	if envId != "" {
		env, err = services.GetEnvById(tx, envId)
		if err != nil && err.Code() == e.EnvNotExists {
			return nil, e.New(err.Code(), err, http.StatusBadRequest)
		} else if err != nil {
			c.Logger().Errorf("error get environment, err %s", err)
			return nil, e.New(e.DBError, err, http.StatusInternalServerError)
		}
	}

	// 模板检查
	tpl, err := services.GetTemplateById(tx, form.Id)
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
	GroupName string `json:"groupName"`
	Creator   string `json:"creator"`
	Summary
}

// SearchPolicy 查询策略列表
func SearchPolicy(c *ctx.ServiceContext, form *forms.SearchPolicyForm) (interface{}, e.Error) {
	query := services.SearchPolicy(c.DB(), form)
	policyResps := make([]PolicyResp, 0)
	p := page.New(form.CurrentPage(), form.PageSize(), form.Order(query))
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
	} else if len(summaries) > 0 {
		sumMap := make(map[string]*services.PolicyScanSummary, len(policyIds))
		for idx, summary := range summaries {
			sumMap[string(summary.Id)+summary.Status] = summaries[idx]
		}
		for idx, policyResp := range policyResps {
			if summary, ok := sumMap[string(policyResp.Id)+common.PolicyStatusPassed]; ok {
				policyResps[idx].Passed = summary.Count
			}
			if summary, ok := sumMap[string(policyResp.Id)+common.PolicyStatusViolated]; ok {
				policyResps[idx].Violated = summary.Count
			}
			if summary, ok := sumMap[string(policyResp.Id)+common.PolicyStatusFailed]; ok {
				policyResps[idx].Failed = summary.Count
			}
			if summary, ok := sumMap[string(policyResp.Id)+common.PolicyStatusSuppressed]; ok {
				policyResps[idx].Suppressed = summary.Count
			}
		}
	}

	return page.PageResp{
		Total:    p.MustTotal(),
		PageSize: p.Size,
		List:     policyResps,
	}, nil
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

	if form.HasKey("tags") {
		attr["tags"] = form.Tags
	}

	if form.HasKey("enabled") {
		attr["enabled"] = form.Enabled
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

type RespPolicyTpl struct {
	models.Template

	// FIXME: "是否开启检测" 当前无标识
	Status string `json:"status" gorm:"column:scan_task_status"`

	PolicyGroups []services.NewPolicyGroup `json:"policyGroups" gorm:"-"`
	Summary
}

func SearchPolicyTpl(c *ctx.ServiceContext, form *forms.SearchPolicyTplForm) (interface{}, e.Error) {
	respPolicyTpls := make([]RespPolicyTpl, 0)
	tplIds := make([]models.Id, 0)
	query := services.SearchPolicyTpl(c.DB(), form.OrgId, form.TplId, form.Q)
	p := page.New(form.CurrentPage(), form.PageSize(), form.Order(query))
	groupM := make(map[models.Id][]services.NewPolicyGroup, 0)
	if err := p.Scan(&respPolicyTpls); err != nil {
		return nil, e.New(e.DBError, err)
	}
	for _, v := range respPolicyTpls {
		tplIds = append(tplIds, v.Id)
	}

	// 根据模板id查询出关联的所有策略组
	groups, err := services.GetPolicyGroupByTplIds(c.DB(), tplIds)
	if err != nil {
		return nil, err
	}
	for _, v := range groups {
		if _, ok := groupM[v.TplId]; !ok {
			groupM[v.TplId] = []services.NewPolicyGroup{v}
			continue
		}
		groupM[v.TplId] = append(groupM[v.TplId], v)
	}

	for index, v := range respPolicyTpls {
		if _, ok := groupM[v.Id]; !ok {
			respPolicyTpls[index].PolicyGroups = []services.NewPolicyGroup{}
			continue
		}
		respPolicyTpls[index].PolicyGroups = groupM[v.Id]
	}

	// 扫描结果统计信息
	if summaries, err := services.PolicyTargetSummary(c.DB(), tplIds, consts.ScopeTemplate); err != nil {
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	} else if len(summaries) > 0 {
		sumMap := make(map[string]*services.PolicyScanSummary, len(tplIds))
		for idx, summary := range summaries {
			sumMap[string(summary.Id)+summary.Status] = summaries[idx]
		}
		for idx, policyResp := range respPolicyTpls {
			if summary, ok := sumMap[string(policyResp.Id)+common.PolicyStatusPassed]; ok {
				respPolicyTpls[idx].Passed = summary.Count
			}
			if summary, ok := sumMap[string(policyResp.Id)+common.PolicyStatusViolated]; ok {
				respPolicyTpls[idx].Violated = summary.Count
			}
			if summary, ok := sumMap[string(policyResp.Id)+common.PolicyStatusFailed]; ok {
				respPolicyTpls[idx].Failed = summary.Count
			}
			if summary, ok := sumMap[string(policyResp.Id)+common.PolicyStatusSuppressed]; ok {
				respPolicyTpls[idx].Suppressed = summary.Count
			}
		}
	}

	return page.PageResp{
		Total:    p.MustTotal(),
		PageSize: p.Size,
		List:     respPolicyTpls,
	}, nil
}

type RespPolicyEnv struct {
	models.EnvDetail

	Status       string                    `json:"status" gorm:"column:scan_task_status"`
	PolicyGroups []services.NewPolicyGroup `json:"policyGroups" gorm:"-"`
	Summary

	// 以下字段不返回
	TaskStatus string `json:"-" gorm:"-"` // 环境部署任务状态
}

func SearchPolicyEnv(c *ctx.ServiceContext, form *forms.SearchPolicyEnvForm) (interface{}, e.Error) {
	respPolicyEnvs := make([]RespPolicyEnv, 0)
	envIds := make([]models.Id, 0)
	query := services.SearchPolicyEnv(c.DB(), form.OrgId, form.ProjectId, form.EnvId, form.Q)
	p := page.New(form.CurrentPage(), form.PageSize(), form.Order(query))
	groupM := make(map[models.Id][]services.NewPolicyGroup, 0)

	if err := p.Scan(&respPolicyEnvs); err != nil {
		return nil, e.New(e.DBError, err)
	}
	for _, v := range respPolicyEnvs {
		envIds = append(envIds, v.Id)
	}

	// 根据环境id查询出关联的所有策略组
	groups, err := services.GetPolicyGroupByEnvIds(c.DB(), envIds)
	if err != nil {
		return nil, err
	}

	for _, v := range groups {
		if _, ok := groupM[v.EnvId]; !ok {
			groupM[v.EnvId] = []services.NewPolicyGroup{v}
			continue
		}
		groupM[v.EnvId] = append(groupM[v.EnvId], v)
	}

	for index, v := range respPolicyEnvs {
		if _, ok := groupM[v.Id]; !ok {
			respPolicyEnvs[index].PolicyGroups = []services.NewPolicyGroup{}
			continue
		}
		respPolicyEnvs[index].PolicyGroups = groupM[v.Id]
	}

	// 扫描结果统计信息
	if summaries, err := services.PolicyTargetSummary(c.DB(), envIds, consts.ScopeEnv); err != nil {
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	} else if len(summaries) > 0 {
		sumMap := make(map[string]*services.PolicyScanSummary, len(envIds))
		for idx, summary := range summaries {
			sumMap[string(summary.Id)+summary.Status] = summaries[idx]
		}
		for idx, policyResp := range respPolicyEnvs {
			if summary, ok := sumMap[string(policyResp.Id)+common.PolicyStatusPassed]; ok {
				respPolicyEnvs[idx].Passed = summary.Count
			}
			if summary, ok := sumMap[string(policyResp.Id)+common.PolicyStatusViolated]; ok {
				respPolicyEnvs[idx].Violated = summary.Count
			}
			if summary, ok := sumMap[string(policyResp.Id)+common.PolicyStatusFailed]; ok {
				respPolicyEnvs[idx].Failed = summary.Count
			}
			if summary, ok := sumMap[string(policyResp.Id)+common.PolicyStatusSuppressed]; ok {
				respPolicyEnvs[idx].Suppressed = summary.Count
			}
		}
	}

	return page.PageResp{
		Total:    p.MustTotal(),
		PageSize: p.Size,
		List:     respPolicyEnvs,
	}, nil
}

type RespEnvOfPolicy struct {
	models.Policy
	GroupName string `json:"groupName"`
	GroupId   string `json:"groupId"`
	EnvName   string `json:"envName"`
}

func EnvOfPolicy(c *ctx.ServiceContext, form *forms.EnvOfPolicyForm) (interface{}, e.Error) {
	resp := make([]RespEnvOfPolicy, 0)
	query := services.EnvOfPolicy(c.DB(), form, c.OrgId, c.ProjectId)
	p := page.New(form.CurrentPage(), form.PageSize(), form.Order(query))
	if err := p.Scan(&resp); err != nil {
		return nil, e.New(e.DBError, err)
	}
	return page.PageResp{
		Total:    p.MustTotal(),
		PageSize: p.Size,
		List:     resp,
	}, nil
}

func ValidEnvOfPolicy(c *ctx.ServiceContext, form *forms.EnvOfPolicyForm) (interface{}, e.Error) {
	policies, err := services.GetPoliciesByEnvId(c.DB(), form.Id)
	if err != nil {
		return nil, err
	}
	return policies, nil
}

type RespTplOfPolicy struct {
	models.Policy
	GroupName string `json:"groupName"`
	GroupId   string `json:"groupId"`
	TplName   string `json:"tplName"`
}

func TplOfPolicy(c *ctx.ServiceContext, form *forms.TplOfPolicyForm) (interface{}, e.Error) {
	resp := make([]RespTplOfPolicy, 0)
	query := services.TplOfPolicy(c.DB(), form, c.OrgId, c.ProjectId)
	p := page.New(form.CurrentPage(), form.PageSize(), form.Order(query))
	if err := p.Scan(&resp); err != nil {
		return nil, e.New(e.DBError, err)
	}
	return page.PageResp{
		Total:    p.MustTotal(),
		PageSize: p.Size,
		List:     resp,
	}, nil
}

func ValidTplOfPolicy(c *ctx.ServiceContext, form *forms.TplOfPolicyForm) (interface{}, e.Error) {
	policies, err := services.GetPoliciesByTemplateId(c.DB(), form.Id)
	if err != nil {
		return getEmptyListResult(form)
	}
	return policies, nil
}

type PolicyErrorResp struct {
	models.PolicyResult
	TargetId     models.Id `json:"targetId"`
	EnvName      string    `json:"envName"`
	TemplateName string    `json:"templateName"`
}

func (PolicyErrorResp) TableName() string {
	return models.PolicyResult{}.TableName()
}

// PolicyError 获取合规错误列表，包含执行错误和合规不通过，排除已经屏蔽的条目
func PolicyError(c *ctx.ServiceContext, form *forms.PolicyErrorForm) (interface{}, e.Error) {
	query := services.PolicyError(c.DB(), form.Id)
	if form.HasKey("q") {
		query = query.Where(fmt.Sprintf("env_name LIKE '%%%s%%' or template_name LIKE '%%%s%%'", form.Q, form.Q))
	}
	return getPage(query, form, PolicyErrorResp{})
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
	ScanTime    *models.Time   `json:"scanTime"`    // 扫描时间
	ScanResults []PolicyResult `json:"scanResults"` // 扫描结果
}

type PolicyResult struct {
	models.PolicyResult
	PolicyName      string `json:"policyName" example:"VPC 安全组规则"`  // 策略名称
	PolicyGroupName string `json:"policyGroupName" example:"安全策略组"` // 策略组名称
	FixSuggestion   string `json:"fixSuggestion" example:"建议您创建一个专有网络..."`
}

func PolicyScanResult(c *ctx.ServiceContext, scope string, form *forms.PolicyScanResultForm) (interface{}, e.Error) {
	c.AddLogField("action", fmt.Sprintf("scan result %s %s", scope, form.Id))
	var (
		scanTask *models.ScanTask
		err      error
	)
	if scope == consts.ScopeEnv {
		scanTask, err = services.GetEnvLastScanTask(c.DB(), form.Id)
	} else if scope == consts.ScopeTemplate {
		scanTask, err = services.GetTplLastScanTask(c.DB(), form.Id)
	} else {
		return nil, e.New(e.InternalError, fmt.Errorf("unknown policy scan result scope '%s'", scope))
	}

	if err != nil {
		if e.IsRecordNotFound(err) {
			return nil, nil
		}
		return nil, e.AutoNew(err, e.DBError)
	}

	query := services.QueryPolicyResult(c.DB(), scanTask.Id)
	if form.SortField() == "" {
		query = query.Order("policy_group_name, policy_name")
	} else {
		query = form.Order(query)
	}
	results := make([]PolicyResult, 0)
	p := page.New(form.CurrentPage(), form.PageSize(), form.Order(query))
	if err := p.Scan(&results); err != nil {
		return nil, e.New(e.DBError, err)
	}

	return page.PageResp{
		Total:    p.MustTotal(),
		PageSize: p.Size,
		List: &ScanResultResp{
			ScanTime:    scanTask.StartAt,
			ScanResults: results,
		},
	}, nil
}

type Summary struct {
	Passed     int `json:"passed"`
	Violated   int `json:"violated"`
	Suppressed int `json:"suppressed"`
	Failed     int `json:"failed"`
}

type Polyline struct {
	Column []string `json:"column,omitempty" example:"08-20,08-21"`
	Value  []int    `json:"value,omitempty" example:"101,103"`
}

type PieChar []PieSector

type PieSector struct {
	Name  string `json:"name" example:"08-20"`
	Value int    `json:"value" example:"10"`
}

type PolicyScanReportResp struct {
	Total            PieChar  `json:"total"`            // 检测结果比例
	TaskScanCount    Polyline `json:"scanCount"`        // 检测源执行次数
	PolicyScanCount  Polyline `json:"policyScanCount"`  // 策略运行趋势
	PolicyPassedRate Polyline `json:"policyPassedRate"` // 检测通过率趋势
}

func PolicyScanReport(c *ctx.ServiceContext, form *forms.PolicyScanReportForm) (*PolicyScanReportResp, e.Error) {
	if !form.HasKey("to") {
		form.To = time.Now()
	}
	if !form.HasKey("from") {
		// 往回 5 天
		y, m, d := form.To.AddDate(0, 0, -15).Date()
		form.From = time.Date(y, m, d, 0, 0, 0, 0, time.Local)
	}
	scanStatus, err := services.GetPolicyScanStatus(c.DB(), form.Id, form.From, form.To, consts.ScopePolicy)
	if err != nil {
		return nil, e.New(err.Code(), err, http.StatusInternalServerError)
	}

	report := PolicyScanReportResp{}
	totalScan := &report.PolicyScanCount
	passedScan := &report.PolicyPassedRate
	totalSummary := Summary{}

	for _, s := range scanStatus {
		d := s.Date[5:10] // 2021-08-08T00:00:00+08:00 => 08-08
		found := false
		for idx := range totalScan.Column {
			if totalScan.Column[idx] == d {
				// 跳过扫描中状态策略
				if s.Status != common.PolicyStatusPending {
					totalScan.Value[idx] += s.Count
				}
				if s.Status == common.PolicyStatusPassed {
					passedScan.Value[idx] = s.Count
				}

				found = true
				break
			}
		}
		if !found {
			totalScan.Column = append(totalScan.Column, d)
			totalScan.Value = append(totalScan.Value, s.Count)

			passedScan.Column = append(passedScan.Column, d)
			if s.Status == common.PolicyStatusPassed {
				passedScan.Value = append(passedScan.Value, s.Count)
			} else {
				passedScan.Value = append(passedScan.Value, 0)
			}
		}

		switch s.Status {
		case common.PolicyStatusPassed:
			totalSummary.Passed += s.Count
		case common.PolicyStatusViolated:
			totalSummary.Violated += s.Count
		case common.PolicyStatusSuppressed:
			totalSummary.Suppressed += s.Count
		case common.PolicyStatusFailed:
			totalSummary.Failed += s.Count
		}
	}
	report.Total = append(report.Total, PieSector{
		Name:  common.PolicyStatusPassed,
		Value: totalSummary.Passed,
	}, PieSector{
		Name:  common.PolicyStatusViolated,
		Value: totalSummary.Violated,
	}, PieSector{
		Name:  common.PolicyStatusSuppressed,
		Value: totalSummary.Suppressed,
	}, PieSector{
		Name:  common.PolicyStatusFailed,
		Value: totalSummary.Failed,
	})

	scanTaskStatus, err := services.GetPolicyScanByDate(c.DB(), form.Id, form.From, form.To)
	if err != nil {
		return nil, e.New(err.Code(), err, http.StatusInternalServerError)
	}
	taskCount := &report.TaskScanCount

	for _, s := range scanTaskStatus {
		d := s.Date[5:10] // 2021-08-08T00:00:00+08:00 => 08-08
		found := false
		for idx := range taskCount.Column {
			if taskCount.Column[idx] == d {
				taskCount.Value[idx] += 1
				found = true
				break
			}
		}
		if !found {
			taskCount.Column = append(taskCount.Column, d)
			taskCount.Value = append(taskCount.Value, 1)
		}
	}

	return &report, nil
}

type PolylinePercent struct {
	Column []string  `json:"column,omitempty" example:"08-20,08-21"`
	Value  []Percent `json:"value,omitempty" example:"0.951,0.962"`
	Passed []int     `json:"-"`
	Total  []int     `json:"-"`
}

type Percent float64 // 百分比，保留1位百分比小数，0.951 = 95.1%

func (n Percent) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%.3f", n)), nil
}

type PolicyTestResp struct {
	Data  interface{} `json:"data" swaggertype:"string" example:"{\n\"accurics\":{\n\"instanceWithNoVpc\":[\n{\n\"Id\":\"alicloud_instance.instance\"\n}\n]\n}\n}"` // 脚本测试输出，json文本
	Error string      `json:"error" example:"1 error occurred: policy.rego:4: rego_parse_error: refs cannot be used for rule\n"`                                    // 脚本执行错误内容
}

func PolicyTest(c *ctx.ServiceContext, form *forms.PolicyTestForm) (*PolicyTestResp, e.Error) {
	c.AddLogField("action", fmt.Sprintf("test template"))

	if _, _, _, err := parseRegoHeader(form.Rego); err != nil {
		return &PolicyTestResp{
			Data:  "",
			Error: fmt.Sprintf("1 error occurred: %s", err.Error()),
		}, nil
	}
	var value interface{}
	if err := json.Unmarshal([]byte(form.Input), &value); err != nil {
		return &PolicyTestResp{
			Data:  "",
			Error: fmt.Sprintf("invalid input %v", err),
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

	if result, err := policy.EngineScan(regoPath, inputPath); err != nil {
		return &PolicyTestResp{
			Data:  "",
			Error: fmt.Sprintf("%s", err),
		}, nil
	} else {
		output, err := json.Marshal(result)
		if err != nil {
			return &PolicyTestResp{
				Data:  "",
				Error: fmt.Sprintf("marshal output %v", err),
			}, nil
		}
		return &PolicyTestResp{
			Data:  string(output),
			Error: "",
		}, nil
	}
}
