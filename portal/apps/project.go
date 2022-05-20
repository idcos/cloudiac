// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package apps

import (
	"cloudiac/common"
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/libs/page"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/models/resps"
	"cloudiac/portal/services"
	"cloudiac/utils"
	"errors"
	"fmt"
	"net/http"
	"path"
	"strings"
)

func CreateProject(c *ctx.ServiceContext, form *forms.CreateProjectForm) (interface{}, e.Error) {
	tx := c.DB().Begin()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	project, err := services.CreateProject(tx, &models.Project{
		Name:        form.Name,
		OrgId:       c.OrgId,
		Description: form.Description,
		CreatorId:   c.UserId,
	})

	if err != nil && err.Code() == e.ProjectAlreadyExists {
		_ = tx.Rollback()
		return nil, e.New(err.Code(), err, http.StatusBadRequest)
	} else if err != nil {
		c.Logger().Errorf("error creating project, err %s", err)
		_ = tx.Rollback()
		return nil, e.AutoNew(err, e.DBError)
	}

	// 检查用户是否属于本组织用户
	for _, userAuth := range form.UserAuthorization {
		if !services.UserHasOrgRole(userAuth.UserId, c.OrgId, "") {
			return nil, e.New(e.BadParam, fmt.Errorf("invalid user"), http.StatusBadRequest)
		}
	}
	// 如果创建人不是超级管理员就把创建人加到项目里面
	if !c.IsSuperAdmin {
		form.UserAuthorization = append(form.UserAuthorization, forms.UserAuthorization{
			UserId: c.UserId,
			Role:   consts.ProjectRoleManager,
		})
	}

	if err := services.BindProjectUsers(tx, project.Id, form.UserAuthorization); err != nil {
		c.Logger().Errorf("error creating project user, err %s", err)
		_ = tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		c.Logger().Errorf("error commit, err %s", err)
		_ = tx.Rollback()
		return nil, e.New(e.DBError, err)
	}
	return project, nil
}

func SearchProject(c *ctx.ServiceContext, form *forms.SearchProjectForm) (interface{}, e.Error) {
	queryStatus := form.Status
	if queryStatus == "" {
		// 默认只查询启用状态的项目
		queryStatus = common.ProjectStatusEnable
	}
	query := services.SearchProject(c.DB(), c.OrgId, form.Q, queryStatus)

	if !c.IsSuperAdmin && !services.UserHasOrgRole(c.UserId, c.OrgId, consts.OrgRoleAdmin) {
		projectIds, err := getSearchProjectIds(query, c.UserId, c.OrgId, form.ProjectId)
		if err != nil {
			c.Logger().Errorf("error get projects, err %s", err)
			return nil, e.New(e.DBError, err)
		}
		if len(projectIds) > 0 {
			query = query.Where(fmt.Sprintf("%s.id in (?)", models.Project{}.TableName()), projectIds)
		} else {
			return getEmptyListResult(form)
		}
	} else {
		if form.ProjectId != "" {
			query = query.Where(fmt.Sprintf("%s.id = ?", models.Project{}.TableName()), form.ProjectId)
		}
	}

	// 默认按创建时间逆序排序
	if form.SortField() == "" {
		query = query.Order("created_at DESC")
	}
	// 查询用户名称
	query = query.Joins(fmt.Sprintf("left join %s as user on user.id = %s.creator_id",
		models.User{}.TableName(), models.Project{}.TableName())).
		LazySelectAppend(fmt.Sprintf("%s.*,user.name as creator", models.Project{}.TableName()))
	p := page.New(form.CurrentPage(), form.PageSize(), query)
	projectResp := make([]resps.ProjectResp, 0)
	if err := p.Scan(&projectResp); err != nil {
		c.Logger().Errorf("error get project info, err %s", err)
		return nil, e.New(e.DBError, err)
	}

	db := c.DB()
	// 活跃的环境数量
	if err := setProjectActiveEnvs(db, projectResp); err != nil {
		return nil, err
	}

	// 是否需要统计数据
	if form.WithStat {
		err := setProjectResStatData(db, projectResp)
		if err != nil {
			return nil, err
		}
	}

	return page.PageResp{
		Total:    p.MustTotal(),
		PageSize: p.Size,
		List:     projectResp,
	}, nil
}

func getSearchProjectIds(query *db.Session, userId, orgId, projectId models.Id) ([]models.Id, e.Error) {
	projectIds, err := services.GetProjectsByUserOrg(query, userId, orgId)
	if err != nil {
		return nil, err
	}

	if projectId == "" {
		return projectIds, nil
	}

	var isExist = false
	for _, id := range projectIds {
		if projectId == id {
			isExist = true
			break
		}
	}

	if !isExist {
		return nil, e.New(e.InvalidProjectId, fmt.Errorf("Can not access this project Id: %s", projectId))
	}

	return []models.Id{projectId}, nil
}

func setProjectResStatData(db *db.Session, projectResp []resps.ProjectResp) e.Error {
	// 参与检索的projects
	searchedProjectIds := make([]models.Id, 0)
	for _, resp := range projectResp {
		searchedProjectIds = append(searchedProjectIds, resp.Id)
	}

	// 获取项目的资源变化趋势
	mResStatData, err := services.GetResGrowTrendByProjects(db, searchedProjectIds, 7)
	if err != nil {
		return err
	}

	// 加入项目的资源变化趋势数据
	for i := range projectResp {
		projectResp[i].ResStats = mResStatData[projectResp[i].Id]
	}

	return nil
}

func setProjectActiveEnvs(db *db.Session, projectResp []resps.ProjectResp) e.Error {
	searchedProjectIds := make([]models.Id, 0)
	for _, resp := range projectResp {
		searchedProjectIds = append(searchedProjectIds, resp.Id)
	}

	m, err := services.GetProjectActiveEnvs(db, searchedProjectIds)
	if err != nil {
		return err
	}

	// 加入项目的活跃环境数量
	for i := range projectResp {
		projectResp[i].ActiveEnvironment = m[projectResp[i].Id]
	}

	return nil
}

func UpdateProject(c *ctx.ServiceContext, form *forms.UpdateProjectForm) (interface{}, e.Error) {
	// 先删除项目和用户关系，在重新创建
	tx := c.DB().Begin()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	//校验用户是否在该项目下有权限
	isExist := IsUserOrgProjectPermission(tx, c.UserId, form.Id, consts.ProjectRoleManager)
	isExistOrg := IsUserOrgPermission(tx, c.UserId, c.OrgId, consts.OrgRoleAdmin)
	if !isExist && !c.IsSuperAdmin && !isExistOrg {
		return nil, e.New(e.ObjectNotExistsOrNoPerm, http.StatusForbidden, errors.New("not permission"))
	}

	//修改项目数据
	attrs := models.Attrs{}
	if form.HasKey("name") {
		attrs["name"] = form.Name
	}

	if form.HasKey("description") {
		attrs["description"] = form.Description
	}

	if form.HasKey("status") {
		if form.Status == "disable" {
			query := services.QueryProjectEnvResource(tx, form.Id)
			query = services.QueryActiveEnv(query)
			activeEnvs, err := services.GetProActiveEnvs(query)
			if err != nil {
				return nil, err
			}
			if len(activeEnvs) > 0 {
				return nil, e.New(e.ProjectHasActiveEnvs,
					fmt.Errorf("project exists active env"))
			}
		}
		attrs["status"] = form.Status
	}

	project := &models.Project{}
	project.Id = form.Id
	err := services.UpdateProject(tx, project, attrs)

	if err != nil && err.Code() == e.ProjectAliasDuplicate {
		_ = tx.Rollback()
		return nil, e.New(err.Code(), err, http.StatusBadRequest)
	} else if err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("error update project, err %s", err)
		return nil, e.New(e.DBError, err)
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return nil, e.New(e.DBError, err)
	}
	return nil, nil
}

func DeleteProject(c *ctx.ServiceContext, form *forms.DeleteProjectForm) (interface{}, e.Error) {
	return nil, e.New(e.NotImplement)
}

func SearchProjectResourcesFilters(c *ctx.ServiceContext, form *forms.SearchProjectResourceForm) (*resps.OrgEnvAndProviderResp, e.Error) {
	query := services.GetOrgOrProjectResourcesQuery(c.DB().Model(&models.Resource{}), form.Q, c.OrgId, c.ProjectId, c.UserId, c.IsSuperAdmin)
	type SearchResult struct {
		EnvName  string    `json:"env_name"`
		EnvId    models.Id `json:"env_id"`
		Provider string    `json:"provider"`
	}
	rs := make([]SearchResult, 0)
	if err := query.Scan(&rs); err != nil {
		return nil, e.New(e.DBError, err)
	}
	r := &resps.OrgEnvAndProviderResp{}
	temp := map[string]interface{}{}
	for _, v := range rs {
		if _, ok := temp[v.EnvName]; !ok {
			// 通过map 对环境名称进行过滤
			r.Envs = append(r.Envs, resps.EnvResp{EnvName: v.EnvName, EnvId: v.EnvId})
			temp[v.EnvName] = nil
		}
		r.Providers = append(r.Providers, path.Base(v.Provider))
	}
	r.Providers = utils.Set(r.Providers)

	return r, nil
}

func SearchProjectResources(c *ctx.ServiceContext, form *forms.SearchProjectResourceForm) (interface{}, e.Error) {
	query := services.GetOrgOrProjectResourcesQuery(c.DB().Model(&models.Resource{}), form.Q, c.OrgId, c.ProjectId, c.UserId, c.IsSuperAdmin)
	if len(form.EnvIds) != 0 {
		query = query.Where("iac_env.id in (?)", strings.Split(form.EnvIds, ","))
	}
	query = services.GetProviderQuery(form.Providers, query)
	query = query.Order("project_id, env_id, provider desc")
	rs, p, err := services.GetOrgOrProjectResourcesResp(form.CurrentPage(), form.PageSize(), query)
	if err != nil {
		return nil, err
	}
	return &page.PageResp{
		Total:    p.MustTotal(),
		PageSize: p.Size,
		List:     rs,
	}, nil

}

func DetailProject(c *ctx.ServiceContext, form *forms.DetailProjectForm) (interface{}, e.Error) {
	tx := c.DB().Begin()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()
	//校验用户是否在该项目下有权限
	isExist := IsUserOrgProjectPermission(tx, c.UserId, form.Id, consts.ProjectRoleManager)
	isExistOrg := IsUserOrgPermission(tx, c.UserId, c.OrgId, consts.OrgRoleAdmin)
	if !isExist && !isExistOrg && !c.IsSuperAdmin {
		_ = tx.Rollback()
		return nil, e.New(e.ObjectNotExistsOrNoPerm, http.StatusForbidden, errors.New("not permission"))
	}
	projectUser, err := services.SearchProjectUsers(tx, form.Id)
	if err != nil {
		_ = tx.Rollback()
		return nil, e.New(e.DBError, err)
	}
	project, err := services.DetailProject(tx, form.Id)
	if err != nil {
		_ = tx.Rollback()
		return nil, e.New(e.DBError, err)
	}

	tplCount, er := services.StatisticalProjectTpl(tx, form.Id)
	if er != nil {
		_ = tx.Rollback()
		return nil, e.New(e.DBError, er)
	}
	envResp, er := services.StatisticalProjectEnv(tx, form.Id)
	if er != nil {
		_ = tx.Rollback()
		return nil, e.New(e.DBError, er)
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return nil, e.New(e.DBError, err)
	}

	return resps.DetailProjectResp{
		Project:           project,
		UserAuthorization: projectUser,
		ProjectStatistics: resps.ProjectStatistics{
			TplCount:    tplCount,
			EnvActive:   envResp.EnvActive,
			EnvFailed:   envResp.EnvFailed,
			EnvInactive: envResp.EnvInactive,
		},
	}, nil
}

func IsUserOrgPermission(dbSess *db.Session, userId, orgId models.Id, role string) bool {
	isExists, err := services.GetUserRoleByOrg(dbSess, userId, orgId, role)
	if err != nil {
		return isExists
	}
	return isExists
}

func IsUserOrgProjectPermission(dbSess *db.Session, userId, project models.Id, role string) bool {
	isExists, err := services.GetUserRoleByProject(dbSess, userId, project, role)
	if err != nil {
		return isExists
	}
	return isExists
}

// ProjectStat 组织和项目概览页统计数据
func ProjectStat(c *ctx.ServiceContext, form *forms.ProjectStatForm) (interface{}, e.Error) {
	tx := c.DB()
	// 环境状态占比
	envStat, err := services.GetProjectEnvStat(tx, form.ProjectId)
	if err != nil {
		return nil, err
	}

	// 资源类型占比
	resStat, err := services.GetProjectResStat(tx, form.ProjectId, form.Limit)
	if err != nil {
		return nil, err
	}

	// 环境资源数量
	envResStat, err := services.GetProjectEnvResStat(tx, form.ProjectId, form.Limit)
	if err != nil {
		return nil, err
	}

	// 资源新增趋势
	resGrowTrend, err := services.GetProjectResGrowTrend(tx, form.ProjectId, 7)
	if err != nil {
		return nil, err
	}

	return &resps.ProjectStatResp{
		EnvStat:      envStat,
		ResStat:      resStat,
		EnvResStat:   envResStat,
		ResGrowTrend: resGrowTrend,
	}, nil
}
