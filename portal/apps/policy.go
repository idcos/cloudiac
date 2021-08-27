package apps

import (
	"cloudiac/common"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

// CreatePolicy 创建策略
func CreatePolicy(c *ctx.ServiceContext, form *forms.CreatePolicyForm) (*models.Policy, e.Error) {
	c.AddLogField("action", fmt.Sprintf("create policy %s", form.Name))

	entryName, resourceType, policyType, err := parseRegoHeader(form.Rego)
	if err != nil {
		return nil, e.New(err.Code(), err, http.StatusBadRequest)
	}

	p := models.Policy{
		Name:          form.Name,
		CreatorId:     c.UserId,
		FixSuggestion: form.FixSuggestion,
		Severity:      form.Severity,
		Rego:          form.Rego,
		Entry:         entryName,
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
func parseRegoHeader(rego string) (entry string, policyType string, resType string, err e.Error) {
	regex := regexp.MustCompile("(?m)^##entry (.*)$")
	match := regex.FindStringSubmatch(rego)
	if len(match) == 2 {
		entry = strings.TrimSpace(match[1])
	} else {
		return "", "", "", e.New(e.PolicyRegoMissingComment)
	}

	regex = regexp.MustCompile("(?m)^##policyType (.*)$")
	match = regex.FindStringSubmatch(rego)
	if len(match) == 2 {
		policyType = strings.TrimSpace(match[1])
	} else {
		return "", "", "", e.New(e.PolicyRegoMissingComment)
	}

	regex = regexp.MustCompile("(?m)^##resourceType (.*)$")
	match = regex.FindStringSubmatch(rego)
	if len(match) == 2 {
		resType = strings.TrimSpace(match[1])
	} else {
		return "", "", "", e.New(e.PolicyRegoMissingComment)
	}

	return
}

// ScanTemplate 扫描云模板策略
func ScanTemplate(c *ctx.ServiceContext, form *forms.ScanTemplateForm) (*models.ScanTask, e.Error) {
	c.AddLogField("action", fmt.Sprintf("create scan template %s", form.Id))
	if c.OrgId == "" {
		return nil, e.New(e.BadRequest, http.StatusBadRequest)
	}
	// TODO: 检查云模板访问权限

	tx := c.Tx()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	// 模板检查
	tplQuery := services.QueryWithOrgId(tx, c.OrgId)
	tpl, err := services.GetTemplateById(tplQuery, form.Id)
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
	task, err := services.CreateScanTask(tx, tpl, nil, models.ScanTask{
		Name:      models.ScanTask{}.GetTaskNameByType(models.TaskTypeScan),
		CreatorId: c.UserId,
		Revision:  tpl.RepoRevision,
		BaseTask: models.BaseTask{
			Type:        models.TaskTypeScan,
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
