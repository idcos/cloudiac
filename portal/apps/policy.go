// Copyright 2021 CloudJ Company Limited. All rights reserved.

package apps

import (
	"cloudiac/common"
	"cloudiac/policy"
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/libs/page"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
	"cloudiac/portal/services/logstorage"
	"cloudiac/utils"
	"cloudiac/utils/logs"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
)

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

// 发起执行多个云模版的合规检测任务
func ScanTemplates(c *ctx.ServiceContext, form *forms.ScanTemplateForms) ([]*models.ScanTask, e.Error) {
	scanTasks := []*models.ScanTask{}
	for _, v := range form.Ids {
		tplForm := &forms.ScanTemplateForm{
			Id: v,
		}
		scanTask, err := ScanTemplateOrEnv(c, tplForm, "")
		if err != nil {
			return nil, err
		}
		scanTasks = append(scanTasks, scanTask)
	}
	return scanTasks, nil
}

// ScanTemplateOrEnv 扫描云模板或环境的合规策略
func ScanTemplateOrEnv(c *ctx.ServiceContext, form *forms.ScanTemplateForm, envId models.Id) (*models.ScanTask, e.Error) {
	c.AddLogField("action", fmt.Sprintf("scan template %s", form.Id))
	if envId != "" {
		c.AddLogField("envId", envId.String())
	}

	tx := c.Tx()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	var (
		tpl       *models.Template
		env       *models.Env
		err       e.Error
		projectId models.Id
	)

	if envId != "" { // 环境检查
		env, err = services.GetEnvById(tx, envId)
		if err != nil && err.Code() == e.EnvNotExists {
			_ = tx.Rollback()
			return nil, e.New(err.Code(), err, http.StatusBadRequest)
		} else if err != nil {
			_ = tx.Rollback()
			c.Logger().Errorf("error get environment, err %s", err)
			return nil, e.New(e.DBError, err, http.StatusInternalServerError)
		}
		projectId = env.ProjectId

		// 环境扫描未启用，不允许发起手动检测
		if enabled, err := services.IsEnvEnabledScan(tx, envId); err != nil {
			_ = tx.Rollback()
			return nil, e.New(e.DBError, err, http.StatusInternalServerError)
		} else if !enabled && !form.Parse {
			_ = tx.Rollback()
			return nil, e.New(e.PolicyScanNotEnabled, http.StatusBadRequest)
		}
	}

	// 模板检查
	tpl, err = services.GetTemplateById(tx, form.Id)
	if err != nil && err.Code() == e.TemplateNotExists {
		_ = tx.Rollback()
		return nil, e.New(err.Code(), err, http.StatusBadRequest)
	} else if err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("error get template, err %s", err)
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	}
	if envId == "" {
		// 云模板扫描未启用，不允许发起手动检测
		if enabled, err := services.IsTemplateEnabledScan(tx, form.Id); err != nil {
			_ = tx.Rollback()
			return nil, e.New(e.DBError, err, http.StatusInternalServerError)
		} else if !enabled && !form.Parse {
			_ = tx.Rollback()
			return nil, e.New(e.PolicyScanNotEnabled, http.StatusBadRequest)
		}
	}

	// 创建任务
	runnerId, err := services.GetDefaultRunnerId()
	if err != nil {
		_ = tx.Rollback()
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
		ProjectId: projectId,
		BaseTask: models.BaseTask{
			Type:        taskType,
			StepTimeout: common.DefaultTaskStepTimeout,
			RunnerId:    runnerId,
		},
	})
	if err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("error creating scan task, err %s", err)
		return nil, e.New(err.Code(), err, http.StatusInternalServerError)
	}

	if task.Type == models.TaskTypeScan {
		if envId != "" { // 环境检查
			env.LastScanTaskId = task.Id
			if _, err := tx.Save(env); err != nil {
				_ = tx.Rollback()
				c.Logger().Errorf("save env, err %s", err)
				return nil, e.New(e.DBError, err, http.StatusInternalServerError)
			}
		} else { // 模板检查
			tpl.LastScanTaskId = task.Id
			if _, err := tx.Save(tpl); err != nil {
				_ = tx.Rollback()
				c.Logger().Errorf("save template, err %s", err)
				return nil, e.New(e.DBError, err, http.StatusInternalServerError)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("commit env, err %s", err)
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

	return ScanTemplateOrEnv(c, &f, env.Id)
}

type PolicyResp struct {
	models.Policy
	GroupName string `json:"groupName"`
	Creator   string `json:"creator"`
	Summary
}

// SearchPolicy 查询策略列表
func SearchPolicy(c *ctx.ServiceContext, form *forms.SearchPolicyForm) (interface{}, e.Error) {
	query := services.SearchPolicy(c.DB(), form, c.OrgId)
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
	if summaries, err := services.PolicySummary(c.DB(), policyIds, consts.ScopePolicy, c.OrgId); err != nil {
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

// DetailPolicy 查询策略组详情
func DetailPolicy(c *ctx.ServiceContext, form *forms.DetailPolicyForm) (interface{}, e.Error) {
	return services.DetailPolicy(c.DB(), form.Id)
}

type RespPolicyTpl struct {
	models.Template

	PolicyStatus string `json:"policyStatus"` // 策略检查状态, enum('passed','violated','pending','failed')

	PolicyGroups []services.NewPolicyGroup `json:"policyGroups" gorm:"-"`
	OrgName      string                    `json:"orgName" form:"orgName" `
	Summary

	// 以下字段不返回
	Status string `json:"status,omitempty" gorm:"-" swaggerignore:"true"` // 模板状态(enabled, disable)
}

func SearchPolicyTpl(c *ctx.ServiceContext, form *forms.SearchPolicyTplForm) (interface{}, e.Error) {
	respPolicyTpls := make([]RespPolicyTpl, 0)
	tplIds := make([]models.Id, 0)
	query := services.SearchPolicyTpl(c.DB(), c.OrgId, form.TplId, form.Q)
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

	// 最后一次扫描结果
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

	PolicyStatus string `json:"policyStatus"` // 策略检查状态, enum('passed','violated','pending','failed')

	PolicyGroups []services.NewPolicyGroup `json:"policyGroups" gorm:"-"`
	Summary
	OrgName     string `json:"orgName" form:"orgName" `
	ProjectName string `json:"projectName" form:"projectName" `

	// 以下字段不返回
	Status     string `json:"status,omitempty" gorm:"-" swaggerignore:"true"`     // 环境状态
	TaskStatus string `json:"taskStatus,omitempty" gorm:"-" swaggerignore:"true"` // 环境部署任务状态
}

func SearchPolicyEnv(c *ctx.ServiceContext, form *forms.SearchPolicyEnvForm) (interface{}, e.Error) {
	respPolicyEnvs := make([]RespPolicyEnv, 0)
	envIds := make([]models.Id, 0)
	query := services.SearchPolicyEnv(c.DB(), c.OrgId, form.ProjectId, form.EnvId, form.Q)
	p := page.New(form.CurrentPage(), form.PageSize(), form.Order(query))
	groupM := make(map[models.Id][]services.NewPolicyGroup)

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

type RespTplOfPolicyGroup struct {
	GroupName string `json:"groupName"`
	GroupId   string `json:"groupId"`
}

func TplOfPolicyGroup(c *ctx.ServiceContext, form *forms.TplOfPolicyGroupForm) (interface{}, e.Error) {
	resp := make([]RespTplOfPolicyGroup, 0)
	query := services.TplOfPolicyGroup(c.DB(), form)
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

// PolicyError 获取合规错误列表，包含最后一次检测错误和合规不通过，排除已经屏蔽的条目
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
	scanTask, err := ScanTemplateOrEnv(c, &f, envId)
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

type ScanResultPageResp struct {
	Task     *models.ScanTask `json:"task"`     // 扫描任务
	Total    int64            `json:"total"`    // 总数
	PageSize int              `json:"pageSize"` // 分页数量
	List     []PolicyResult   `json:"list"`     // 扫描结果
}

type PolicyResult struct {
	models.PolicyResult
	PolicyName      string `json:"policyName" example:"VPC 安全组规则"`  // 策略名称
	PolicyGroupName string `json:"policyGroupName" example:"安全策略组"` // 策略组名称
	FixSuggestion   string `json:"fixSuggestion" example:"建议您创建一个专有网络..."`
	Rego            string `json:"rego" example:""` //rego 代码文件内容

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

	return ScanResultPageResp{
		Task:     scanTask,
		Total:    p.MustTotal(),
		PageSize: p.Size,
		List:     results,
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
	Total            PieChar         `json:"total"`            // 检测结果比例
	TaskScanCount    Polyline        `json:"scanCount"`        // 检测源执行次数
	PolicyScanCount  Polyline        `json:"policyScanCount"`  // 策略运行趋势
	PolicyPassedRate PolylinePercent `json:"policyPassedRate"` // 检测通过率趋势
}

func PolicyScanReport(c *ctx.ServiceContext, form *forms.PolicyScanReportForm) (*PolicyScanReportResp, e.Error) {
	if !form.HasKey("showCount") {
		// 默认展示最近五个
		form.ShowCount = 5
	}
	if !form.HasKey("to") {
		form.To = time.Now()
	}
	if !form.HasKey("from") {
		// 默认展示近 5 天的数据
		form.From = utils.LastDaysMidnight(5, form.To)
	}

	timePoints := make([]string, 0)
	{
		start := form.From
		for !start.After(form.To) {
			_, m, d := start.Date()
			timePoints = append(timePoints, fmt.Sprintf("%02d-%02d", m, d))
			start = start.AddDate(0, 0, 1)
		}
	}

	scanStatus, err := services.GetPolicyScanStatus(c.DB(), form.Id, form.From, form.To, consts.ScopePolicy)
	if err != nil {
		return nil, e.New(err.Code(), err, http.StatusInternalServerError)
	}

	report := PolicyScanReportResp{}
	totalScan := &report.PolicyScanCount
	passedScan := &report.PolicyPassedRate
	totalSummary := Summary{}

	// 初始化返回的时间点序列，保证没有查询到数据的时候点也会自动填充 0 值
	for _, d := range timePoints {
		totalScan.Column = append(totalScan.Column, d)
		totalScan.Value = append(totalScan.Value, 0)

		passedScan.Column = append(passedScan.Column, d)
		passedScan.Value = append(passedScan.Value, 0)
	}

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
					passedScan.Value[idx] = Percent(s.Count)
				}

				found = true
				break
			}
		}
		if !found {
			c.Logger().Warnf("date '%s' not in time range %v", d, timePoints)
			return nil, e.New(e.InternalError, fmt.Errorf("date '%s' not in time range", d))
		}

		for idx := range totalScan.Column {
			if totalScan.Value[idx] == 0 {
				passedScan.Value[idx] = 0
				continue
			}
			passedScan.Value[idx] = passedScan.Value[idx] / Percent(totalScan.Value[idx])
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

	scanTaskStatus, err := services.GetPolicyScanByTarget(c.DB(), form.Id, form.From, form.To, form.ShowCount)
	if err != nil {
		return nil, e.New(err.Code(), err, http.StatusInternalServerError)
	}
	taskCount := &report.TaskScanCount

	for _, s := range scanTaskStatus {
		taskCount.Column = append(taskCount.Column, s.Name)
		taskCount.Value = append(taskCount.Value, s.Count)
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
	c.AddLogField("action", "test template")

	if _, _, _, err := parseRegoHeader(form.Rego); err != nil {
		return &PolicyTestResp{
			Data:  map[string]interface{}{},
			Error: fmt.Sprintf("1 error occurred: %s", err.Error()),
		}, nil
	}
	var value interface{}
	if err := json.Unmarshal([]byte(form.Input), &value); err != nil {
		return &PolicyTestResp{
			Data:  map[string]interface{}{},
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
			Data:  map[string]interface{}{},
			Error: fmt.Sprintf("%s", err),
		}, nil
	} else {
		return &PolicyTestResp{
			Data:  result,
			Error: "",
		}, nil
	}
}

type PieCharPercent []PieSectorPercent

type PieSectorPercent struct {
	Name  string  `json:"name" example:"passed"`
	Value float64 `json:"value" example:"0.2"`
}

type PolicySummaryResp struct {
	ActivePolicy struct {
		Total   int     `json:"total"`   // 最近 15 天产生扫描记录的策略数量
		Last    int     `json:"last"`    // 16～30 天产生扫描记录的策略数量
		Changes float64 `json:"changes"` // 相较上次变化
		Summary PieChar `json:"summary"` // 策略状态包含
	} `json:"activePolicy"`

	UnresolvedPolicy struct {
		Total   int     `json:"total"`   // 最近 15 天产生扫描记录的策略数量
		Last    int     `json:"last"`    // 16～30 天产生扫描记录的策略数量
		Changes float64 `json:"changes"` // 相较上次变化： (total - last) / last
		Summary PieChar `json:"summary"` // 策略严重级别统计
	} `json:"unresolvedPolicy"`

	PolicyViolated      PieChar `json:"policyViolated"`      // 策略不通过
	PolicyGroupViolated PieChar `json:"policyGroupViolated"` // 策略组不通过
}

func PolicySummary(c *ctx.ServiceContext) (*PolicySummaryResp, e.Error) {
	// 策略概览
	// 默认统计时间范围：最近15天
	// 1. 活跃策略
	//    活跃策略定义：产生扫描记录的策略
	//    last （最近15天）定义：第前30～16天
	//    changes 定义：(total - last) / last
	//    summary: 含 passed / violated / failed / suppressed
	// 2. 未解决错误策略
	//    未解决错误定义：策略在任意一个环境或模板上最后一个扫描结果为 violated 或者 failed
	//    扇形图：按策略的严重级别进行统计
	// 3. 策略检测未通过
	//    未通过定义：策略扫描结果为 violated
	//    柱状图：按未通过次数统计，以策略为纬度，取未通过次数最多的 5 条策略记录
	// 4. 策略组检测未通过
	//    未通过定义：扫描结果含 violated 的策略组
	//    柱状图：按未通过次数统计，以策略组为纬度，取未通过次数最多的 5 条策略组记录

	// 最近 15 天
	to := time.Now()
	from := utils.LastDaysMidnight(15, to)
	// 前 16～30 天
	lastTo := from
	lastFrom := utils.LastDaysMidnight(15, lastTo)

	dbSess := c.DB()
	summaryResp := PolicySummaryResp{}

	// 近15天数据
	scanStatus, err := services.GetPolicyStatusByPolicy(dbSess, from, to, "", c.OrgId)
	if err != nil {
		return nil, e.New(err.Code(), err, http.StatusInternalServerError)
	}
	totalPolicyMap := make(map[models.Id]int)
	policyStatusMap := make(map[string]int)
	for _, v := range scanStatus {
		// 计算策略数量
		totalPolicyMap[v.Id] = 1
		// 按状态统计数量
		policyStatusMap[v.Status] += v.Count

		// // 计算未通过错误策略数量
		// if v.Status == common.PolicyStatusViolated {
		// 	unresolvedCountMap[v.Id]++
		// }
	}

	// 16～30天数据
	lastScanStatus, err := services.GetPolicyStatusByPolicy(dbSess, lastFrom, lastTo, "", c.OrgId)
	if err != nil {
		return nil, e.New(err.Code(), err, http.StatusInternalServerError)
	}
	lastPolicyMap := make(map[models.Id]int)
	for _, v := range lastScanStatus {
		// 计算策略数量
		lastPolicyMap[v.Id] = 1
	}

	// 1. 活跃策略
	// c.Logger().Errorf("totalPolicyMap %+v", totalPolicyMap)
	summaryResp.ActivePolicy.Total = len(totalPolicyMap)
	summaryResp.ActivePolicy.Last = len(lastPolicyMap)
	if summaryResp.ActivePolicy.Last != 0 {
		summaryResp.ActivePolicy.Changes =
			(float64(summaryResp.ActivePolicy.Total) - float64(summaryResp.ActivePolicy.Last)) /
				float64(summaryResp.ActivePolicy.Last)
	} else {
		summaryResp.ActivePolicy.Changes = 1
	}

	s := PieChar{}
	s = append(s, PieSector{
		Name:  common.PolicyStatusPassed,
		Value: policyStatusMap[common.PolicyStatusPassed],
	}, PieSector{
		Name:  common.PolicyStatusViolated,
		Value: policyStatusMap[common.PolicyStatusViolated],
	}, PieSector{
		Name:  common.PolicyStatusFailed,
		Value: policyStatusMap[common.PolicyStatusFailed],
	}, PieSector{
		Name:  common.PolicyStatusSuppressed,
		Value: policyStatusMap[common.PolicyStatusSuppressed],
	})
	summaryResp.ActivePolicy.Summary = s

	// 2. 未解决错误策略
	{
		// 最近 15 天
		unresolvedPolicies, err := services.QueryPolicyStatusEveryTargetLastRun(dbSess, from, to)
		if err != nil {
			return nil, e.AutoNew(errors.Wrap(err, "QueryPolicyStatusEveryTargetLastRun"), e.DBError)
		}

		// 前一个 15天
		lastUnresolvedPolicies, err := services.QueryPolicyStatusEveryTargetLastRun(dbSess, lastFrom, lastTo)
		if err != nil {
			return nil, e.AutoNew(errors.Wrap(err, "QueryPolicyStatusEveryTargetLastRun"), e.DBError)
		}

		summaryResp.UnresolvedPolicy.Total = len(unresolvedPolicies)
		summaryResp.UnresolvedPolicy.Last = len(lastUnresolvedPolicies)
		if summaryResp.UnresolvedPolicy.Last != 0 {
			summaryResp.UnresolvedPolicy.Changes =
				(float64(summaryResp.UnresolvedPolicy.Total) -
					float64(summaryResp.UnresolvedPolicy.Last)) /
					float64(summaryResp.UnresolvedPolicy.Last)
		} else {
			summaryResp.UnresolvedPolicy.Changes = 1
		}
		var high, medium, low int
		for _, v := range unresolvedPolicies {
			switch v.Severity {
			case common.PolicySeverityHigh:
				high++
			case common.PolicySeverityMedium:
				medium++
			case common.PolicySeverityLow:
				low++
			}
		}
		s = PieChar{}
		s = append(s, PieSector{
			Name:  common.PolicySeverityHigh,
			Value: high,
		}, PieSector{
			Name:  common.PolicySeverityMedium,
			Value: medium,
		}, PieSector{
			Name:  common.PolicySeverityLow,
			Value: low,
		})
		summaryResp.UnresolvedPolicy.Summary = s
	}

	// 3. 策略未通过
	violatedScanStatus, err := services.GetPolicyStatusByPolicy(dbSess, from, to, common.PolicyStatusViolated, c.OrgId)
	if err != nil {
		return nil, e.New(err.Code(), err, http.StatusInternalServerError)
	}
	p := PieChar{}
	for i := 0; i < 5 && i < len(violatedScanStatus); i++ {
		p = append(p, PieSector{
			Name:  violatedScanStatus[i].Name,
			Value: violatedScanStatus[i].Count,
		})
	}
	summaryResp.PolicyViolated = p

	// 4. 策略组未通过
	violatedGroupScanStatus, err := services.GetPolicyStatusByPolicyGroup(dbSess, from, to, common.PolicyStatusViolated)
	if err != nil {
		return nil, e.New(err.Code(), err, http.StatusInternalServerError)
	}
	p = PieChar{}
	for i := 0; i < 5 && i < len(violatedGroupScanStatus); i++ {
		p = append(p, PieSector{
			Name:  violatedGroupScanStatus[i].Name,
			Value: violatedGroupScanStatus[i].Count,
		})
	}
	summaryResp.PolicyGroupViolated = p

	return &summaryResp, nil
}

func RepoPolicyCreate(tx *db.Session, g *models.PolicyGroup, orgId, userId models.Id) e.Error {
	// 策略仓库同步
	// 1. 生成临时工作目录
	logger := logs.Get()
	logger.Debugf("creating tmp dir")
	tmpDir, er := os.MkdirTemp("", "*")
	if er != nil {
		return e.New(e.InternalError, errors.Wrapf(er, "create tmp dir"), http.StatusInternalServerError)
	}
	//defer os.RemoveAll(tmpDir)
	logger.Debugf("tmp dir %s", tmpDir)

	// 2. clone 策略组
	var wg sync.WaitGroup
	result := services.DownloadPolicyGroupResult{Group: g}
	wg.Add(1)
	go services.DownloadPolicyGroup(db.Get(), tmpDir, &result, &wg)
	wg.Wait()
	if result.Error != nil {
		logger.Errorf("error download policy group, err %s", result.Error)
		return result.Error
	}

	// 3. 遍历策略组目录，解析策略文件
	logger.Debugf("parsing policy group")
	// TODO: import summary
	policies, er := services.ParsePolicyGroup(filepath.Join(tmpDir, "code", g.Dir))
	if er != nil {
		return e.New(e.InternalError, errors.Wrapf(er, "parse rego"), http.StatusInternalServerError)
	}

	// 策略文件入库
	logger.Debugf("creating policy")
	for _, p := range policies {
		p := models.Policy{
			OrgId:     orgId,
			CreatorId: userId,
			GroupId:   g.Id,

			Name:          p.Meta.Name,
			RuleName:      p.Meta.Name,
			ReferenceId:   p.Meta.ReferenceId,
			Revision:      p.Meta.Version,
			FixSuggestion: p.Meta.FixSuggestion,
			Severity:      p.Meta.Severity,
			ResourceType:  p.Meta.ResourceType,
			PolicyType:    p.Meta.PolicyType,
			Tags:          p.Meta.Category,

			Rego: p.Rego,
		}
		_, er = tx.Save(&p)
		if er != nil {
			return e.New(e.DBError, er)
		}
	}

	return nil
}
