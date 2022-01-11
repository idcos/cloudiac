// Copyright 2021 CloudJ Company Limited. All rights reserved.

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
	"cloudiac/utils"
	"fmt"
	"github.com/Masterminds/semver"
	"github.com/pkg/errors"
	"net/http"
	"strings"
	"time"
)

// CreatePolicyGroup 创建策略组
func CreatePolicyGroup(c *ctx.ServiceContext, form *forms.CreatePolicyGroupForm) (*models.PolicyGroup, e.Error) {
	c.AddLogField("action", fmt.Sprintf("create policy group %s", form.Name))
	logger := c.Logger()
	g := models.PolicyGroup{
		Name:        form.Name,
		Description: form.Description,
		Label:       strings.Join(form.Labels, ","),
		Source:      form.Source,
		VcsId:       form.VcsId,
		RepoId:      form.RepoId,
		OrgId:       c.OrgId,
		CreatorId:   c.UserId,
	}

	if form.HasKey("gitTags") && form.GitTags != "" {
		g.GitTags = form.GitTags
		// 检查是否有效的语义话版本
		v, err := semver.NewVersion(g.GitTags)
		if err != nil {
			return nil, e.AutoNew(fmt.Errorf("git tag is invalid semver"), e.BadParam, http.StatusBadRequest)
		}
		g.Version = v.String()
	} else if form.HasKey("branch") && form.Branch != "" {
		g.Branch = form.Branch
		g.UseLatest = true
	} else {
		return nil, e.New(e.BadParam, http.StatusBadRequest)
	}
	if form.HasKey("dir") {
		g.Dir = form.Dir
	} else {
		g.Dir = consts.DirRoot
	}

	// 策略组仓库解析
	policies, er := PolicyGroupRepoDownloadAndParse(&g)
	if er != nil {
		return nil, e.New(e.InternalError, errors.Wrapf(er, "parse rego"), http.StatusInternalServerError)
	}

	tx := c.Tx()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	// 策略组创建
	group, err := services.CreatePolicyGroup(tx, &g)
	if err != nil && err.Code() == e.PolicyGroupAlreadyExist {
		_ = tx.Rollback()
		return nil, e.New(err.Code(), err, http.StatusBadRequest)
	} else if err != nil {
		_ = tx.Rollback()
		logger.Errorf("error creating policy group, err %s", err)
		return nil, e.AutoNew(err, http.StatusInternalServerError, e.DBError)
	}

	// 策略创建
	err = policiesUpsert(tx, c.UserId, c.OrgId, group, policies)
	if err != nil {
		return nil, e.AutoNew(err, http.StatusInternalServerError, e.DBError)
	}

	if err := tx.Commit(); err != nil {
		logger.Errorf("error commit policy group, err %s", err)
		_ = tx.Rollback()
		return nil, e.New(e.DBError, err)
	}

	return group, nil
}

type PolicyGroupResp struct {
	models.PolicyGroup
	PolicyCount uint     `json:"policyCount" example:"10"`
	RelCount    uint     `json:"relCount"`
	Labels      []string `json:"labels" gorm:"-"`
}

// SearchPolicyGroup 查询策略组列表
func SearchPolicyGroup(c *ctx.ServiceContext, form *forms.SearchPolicyGroupForm) (interface{}, e.Error) {
	query := services.QueryWithOrgId(c.DB(), c.OrgId)
	query = services.SearchPolicyGroup(query, c.OrgId, form.Q)
	policyGroupResps := make([]PolicyGroupResp, 0)
	p := page.New(form.CurrentPage(), form.PageSize(), form.Order(query))
	if err := p.Scan(&policyGroupResps); err != nil {
		return nil, e.New(e.DBError, err)
	}

	for index, pg := range policyGroupResps {
		labels := make([]string, 0)
		if pg.Label != "" {
			labels = strings.Split(pg.Label, ",")
		}
		policyGroupResps[index].Labels = labels
	}
	return page.PageResp{
		Total:    p.MustTotal(),
		PageSize: p.Size,
		List:     policyGroupResps,
	}, nil
}

// UpdatePolicyGroup 修改策略组
func UpdatePolicyGroup(c *ctx.ServiceContext, form *forms.UpdatePolicyGroupForm) (interface{}, e.Error) {
	attr := models.Attrs{}
	if form.HasKey("name") {
		attr["name"] = form.Name
	}

	if form.HasKey("description") {
		attr["description"] = form.Description
	}

	if form.HasKey("enabled") {
		attr["enabled"] = form.Enabled
	}

	if form.HasKey("labels") {
		attr["label"] = strings.Join(form.Labels, ",")
	}

	if form.HasKey("source") {
		attr["source"] = form.Source
	}

	if form.HasKey("vcsId") {
		attr["vcsId"] = form.VcsId
	}

	if form.HasKey("repoId") {
		attr["repoId"] = form.RepoId
	}

	if form.HasKey("gitTags") {
		attr["gitTags"] = form.GitTags
	}

	if form.HasKey("branch") {
		attr["branch"] = form.Branch
	}

	if form.HasKey("dir") {
		if form.Dir == "" {
			attr["dir"] = form.Dir
		} else {
			attr["dir"] = consts.DirRoot
		}
	}

	pg := models.PolicyGroup{}
	pg.Id = form.Id

	var (
		policies []*services.PolicyWithMeta
		er       error
	)
	// 未对仓库信息进行修改时，不重新同步策略数据
	needsSync := false
	if form.HasKey("vcsId") && form.HasKey("repoId") &&
		(form.HasKey("gitTags") || form.HasKey("branch")) && form.HasKey("dir") {
		needsSync = true
		// 策略组仓库解析
		policies, er = PolicyGroupRepoDownloadAndParse(&pg)
		if er != nil {
			return nil, e.New(e.InternalError, errors.Wrapf(er, "parse rego"), http.StatusInternalServerError)
		}
	}

	tx := services.QueryWithOrgId(c.Tx(), c.OrgId)
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	if err := services.UpdatePolicyGroup(tx, &pg, attr); err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	if needsSync {
		// 策略同步
		err := policiesUpsert(tx, c.UserId, c.OrgId, &pg, policies)
		if err != nil {
			_ = tx.Rollback()
			return nil, e.AutoNew(err, http.StatusInternalServerError, e.DBError)
		}
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return nil, e.New(e.DBError, err)
	}

	return nil, nil
}

// DeletePolicyGroup 删除策略组
func DeletePolicyGroup(c *ctx.ServiceContext, form *forms.DeletePolicyGroupForm) (interface{}, e.Error) {
	tx := services.QueryWithOrgId(c.Tx(), c.OrgId)
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	// 解除策略与策略组的关系
	if err := services.RemovePoliciesGroupRelation(tx, form.Id); err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	// 删除策略组
	if err := services.DeletePolicyGroup(tx, form.Id); err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	// 删除策略
	if _, err := services.DeletePolicy(tx, form.Id); err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return nil, e.New(e.DBError, err)
	}

	return nil, nil
}

// DetailPolicyGroup 查询策略组详情
func DetailPolicyGroup(c *ctx.ServiceContext, form *forms.DetailPolicyGroupForm) (interface{}, e.Error) {
	query := services.QueryWithOrgId(c.DB(), c.OrgId)
	pg, err := services.DetailPolicyGroup(query, form.Id)
	if err != nil {
		return nil, err
	}
	labels := make([]string, 0)
	if pg.Label != "" {
		labels = strings.Split(pg.Label, ",")
	}

	return PolicyGroupResp{
		PolicyGroup: *pg,
		Labels:      labels,
	}, nil
}

// OpPolicyAndPolicyGroupRel 创建和修改策略和策略组的关系
func OpPolicyAndPolicyGroupRel(c *ctx.ServiceContext, form *forms.OpnPolicyAndPolicyGroupRelForm) (interface{}, e.Error) {
	tx := services.QueryWithOrgId(c.Tx(), c.OrgId)
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	if form.HasKey("addPolicyIds") && len(form.AddPolicyIds) > 0 {
		for _, policyId := range form.AddPolicyIds {
			policy, err := services.GetPolicyById(tx, models.Id(policyId), c.OrgId)
			if err != nil {
				_ = tx.Rollback()
				return nil, err
			}
			// 一个策略只能关联到一个策略组
			if policy.GroupId != "" {
				_ = tx.Rollback()
				return nil, e.New(e.PolicyBelongedToAnotherGroup, fmt.Errorf("policy belonged to another group"), http.StatusBadRequest)
			}
		}
		// 批量更新
		if affected, err := services.UpdatePolicy(tx.Where("id in (?)", form.AddPolicyIds),
			&models.Policy{}, models.Attrs{"group_id": form.PolicyGroupId}); err != nil {
			_ = tx.Rollback()
			return nil, e.New(e.DBError, err, http.StatusInternalServerError)
		} else if int(affected) != len(form.AddPolicyIds) {
			_ = tx.Rollback()
			return nil, e.New(e.DBError, err, http.StatusInternalServerError)
		}
	}

	if form.HasKey("rmPolicyIds") && len(form.RmPolicyIds) > 0 {
		for _, policyId := range form.RmPolicyIds {
			policy, err := services.GetPolicyById(tx, models.Id(policyId), c.OrgId)
			if err != nil {
				_ = tx.Rollback()
				return nil, err
			}
			// 一个策略只能关联到一个策略组
			if policy.GroupId != form.PolicyGroupId {
				_ = tx.Rollback()
				return nil, e.New(e.PolicyBelongedToAnotherGroup, fmt.Errorf("policy belonged to another group"), http.StatusBadRequest)
			}
		}
		// 批量更新
		if affected, err := services.UpdatePolicy(tx.Where("id in (?)", form.RmPolicyIds),
			&models.Policy{}, models.Attrs{"group_id": ""}); err != nil {
			_ = tx.Rollback()
			return nil, e.New(e.DBError, err, http.StatusInternalServerError)
		} else if int(affected) != len(form.RmPolicyIds) {
			_ = tx.Rollback()
			return nil, e.New(e.DBError, err, http.StatusInternalServerError)
		}
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return nil, e.New(e.DBError, err)
	}
	return nil, nil
}

type LastScanTaskResp struct {
	models.ScanTask
	TargetName  string `json:"targetName"`  // 检查目标
	TargetType  string `json:"targetType"`  // 目标类型：环境/模板
	OrgName     string `json:"orgName"`     // 组织名称
	ProjectName string `json:"projectName"` // 项目
	Creator     string `json:"creator"`     // 创建者
	Summary
}

func PolicyGroupScanTasks(c *ctx.ServiceContext, form *forms.PolicyLastTasksForm) (interface{}, e.Error) {
	query := services.GetPolicyGroupScanTasks(c.DB(), form.Id, c.OrgId)

	// 默认按创建时间逆序排序
	if form.SortField() == "" {
		query = query.Order("created_at DESC")
	} else {
		query = form.Order(query)
	}
	p := page.New(form.CurrentPage(), form.PageSize(), form.Order(query))
	tasks := make([]*LastScanTaskResp, 0)
	if err := p.Scan(&tasks); err != nil {
		return nil, e.New(e.DBError, err)
	}

	// 扫描结果统计信息
	var policyIds []models.Id
	for idx := range tasks {
		policyIds = append(policyIds, tasks[idx].Id)
	}
	if summaries, err := services.PolicySummary(c.DB(), policyIds, consts.ScopeTask, c.OrgId); err != nil {
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	} else if summaries != nil && len(summaries) > 0 {
		sumMap := make(map[string]*services.PolicyScanSummary, len(policyIds))
		for idx, summary := range summaries {
			sumMap[string(summary.Id)+summary.Status] = summaries[idx]
		}
		for idx, policyResp := range tasks {
			if summary, ok := sumMap[string(policyResp.Id)+common.PolicyStatusPassed]; ok {
				tasks[idx].Passed = summary.Count
			}
			if summary, ok := sumMap[string(policyResp.Id)+common.PolicyStatusViolated]; ok {
				tasks[idx].Violated = summary.Count
			}
			if summary, ok := sumMap[string(policyResp.Id)+common.PolicyStatusFailed]; ok {
				tasks[idx].Failed = summary.Count
			}
			if summary, ok := sumMap[string(policyResp.Id)+common.PolicyStatusSuppressed]; ok {
				tasks[idx].Suppressed = summary.Count
			}
		}
	}

	return page.PageResp{
		Total:    p.MustTotal(),
		PageSize: p.Size,
		List:     tasks,
	}, nil
}

func SearchGroupOfPolicy(c *ctx.ServiceContext, form *forms.SearchGroupOfPolicyForm) (interface{}, e.Error) {
	resp := make([]models.Policy, 0)
	query := services.QueryWithOrgId(c.DB(), c.OrgId)
	query = services.SearchGroupOfPolicy(query, form.Id, form.IsBind)
	p := page.New(form.CurrentPage(), form.PageSize(), query)
	if err := p.Scan(&resp); err != nil {
		return nil, e.New(e.DBError, err)
	}
	return page.PageResp{
		Total:    p.MustTotal(),
		PageSize: p.Size,
		List:     resp,
	}, nil
}

type PolicyGroupScanReportResp struct {
	PassedRate PolylinePercent `json:"passedRate"` // 检测通过率
}

func PolicyGroupScanReport(c *ctx.ServiceContext, form *forms.PolicyScanReportForm) (*PolicyGroupScanReportResp, e.Error) {
	if !form.HasKey("to") {
		form.To = time.Now()
	}
	if !form.HasKey("from") {
		// 往回 30 天
		form.From = utils.LastDaysMidnight(30, form.To)
	}
	query := services.QueryWithOrgId(c.DB(), c.OrgId)
	scanStatus, err := services.GetPolicyScanStatus(query, form.Id, form.From, form.To, consts.ScopePolicyGroup)
	if err != nil {
		return nil, e.New(err.Code(), err, http.StatusInternalServerError)
	}

	report := PolicyGroupScanReportResp{}
	r := &report.PassedRate

	for _, s := range scanStatus {
		d := s.Date[5:10] // 2021-08-08T00:00:00+08:00 => 08-08
		found := false
		for idx := range r.Column {
			if r.Column[idx] == d {
				if s.Status == common.PolicyStatusPassed {
					r.Passed[idx] = s.Count
				}
				// FIXME: 是否跳过失败和屏蔽的策略？
				r.Total[idx] += s.Count
				r.Value[idx] = Percent(r.Passed[idx]) / Percent(r.Total[idx])
				found = true
				break
			}
		}
		if !found {
			r.Column = append(r.Column, d)
			if s.Status == common.PolicyStatusPassed {
				r.Passed = append(r.Passed, s.Count)
				r.Total = append(r.Total, s.Count)
				r.Value = append(r.Value, 1)
			} else {
				r.Passed = append(r.Passed, 0)
				r.Total = append(r.Total, s.Count)
				r.Value = append(r.Value, 0)
			}
		}
	}

	return &report, nil
}
