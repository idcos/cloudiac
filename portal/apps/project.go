// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package apps

import (
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/libs/page"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
	"errors"
	"fmt"
	"net/http"
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

type ProjectResp struct {
	models.Project
	Creator string `json:"creator" form:"creator" `
}

func SearchProject(c *ctx.ServiceContext, form *forms.SearchProjectForm) (interface{}, e.Error) {
	query := services.SearchProject(c.DB(), c.OrgId, form.Q, form.Status)
	if !c.IsSuperAdmin && !services.UserHasOrgRole(c.UserId, c.OrgId, consts.OrgRoleAdmin) {
		projectIds, err := services.GetProjectsByUserOrg(query, c.UserId, c.OrgId)
		if err != nil {
			c.Logger().Errorf("error get projects, err %s", err)
			return nil, e.New(e.DBError, err)
		}
		if len(projectIds) > 0 {
			query = query.Where(fmt.Sprintf("%s.id in (?)", models.Project{}.TableName()), projectIds)
		} else {
			return getEmptyListResult(form)
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
	projectResp := make([]ProjectResp, 0)
	if err := p.Scan(&projectResp); err != nil {
		c.Logger().Errorf("error get project info, err %s", err)
		return nil, e.New(e.DBError, err)
	}

	return page.PageResp{
		Total:    p.MustTotal(),
		PageSize: p.Size,
		List:     projectResp,
	}, nil
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
	if !isExist && !c.IsSuperAdmin {
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

type DetailProjectResp struct {
	models.Project
	UserAuthorization []models.UserProject `json:"userAuthorization" form:"userAuthorization" ` //用户认证信息
	ProjectStatistics
}

type ProjectStatistics struct {
	TplCount    int64 `json:"tplCount" form:"tplCount" `
	EnvActive   int64 `json:"envActive" form:"envActive" `
	EnvFailed   int64 `json:"envFailed" form:"envFailed" `
	EnvInactive int64 `json:"envInactive" form:"envInactive" `
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

	return DetailProjectResp{
		project,
		projectUser,
		ProjectStatistics{
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
