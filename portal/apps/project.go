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

func CreateProject(c *ctx.ServiceCtx, form *forms.CreateProjectForm) (interface{}, e.Error) {
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
	// 需要把创建人加进来
	form.UserAuthorization = append(form.UserAuthorization, forms.UserAuthorization{
		UserId: c.UserId,
		Role:   consts.ProjectRoleManager,
	})

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

func SearchProject(c *ctx.ServiceCtx, form *forms.SearchProjectForm) (interface{}, e.Error) {
	query := services.SearchProject(c.DB(), c.OrgId, form.Q, form.Status)
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

func UpdateProject(c *ctx.ServiceCtx, form *forms.UpdateProjectForm) (interface{}, e.Error) {
	// 先删除项目和用户关系，在重新创建
	tx := c.DB().Begin()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	//校验用户是否在该项目下有权限
	isExist := IsUserOrgProjectPermission(tx, c.UserId, form.Id, consts.OrgRoleAdmin)
	if !isExist {
		return nil, e.New(e.ObjectNotExistsOrNoPerm, http.StatusForbidden, errors.New("not permission"))
	}

	if err := services.UpdateProjectUsers(tx, form.Id, form.UserAuthorization); err != nil {
		_ = tx.Rollback()
		return nil, err
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

func DeleteProject(c *ctx.ServiceCtx, form *forms.DeleteProjectForm) (interface{}, e.Error) {
	return nil, e.New(e.NotImplement)
	//tx := c.DB().Begin()
	//defer func() {
	//	if r := recover(); r != nil {
	//		_ = tx.Rollback()
	//		panic(r)
	//	}
	//}()
	////todo 检验环境是否活跃
	////项目是逻辑删除，用户和项目的角色关系是直接删除
	//if err := services.DeleteProject(tx, form.Id); err != nil {
	//	_ = tx.Rollback()
	//	return nil, err
	//}
	//
	//if err := services.DeleteUserProject(tx, form.Id); err != nil {
	//	_ = tx.Rollback()
	//	return nil, err
	//}
	//
	//if err := tx.Commit(); err != nil {
	//	_ = tx.Rollback()
	//	return nil, e.New(e.DBError, err)
	//}
	//
	//return nil, nil
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

func DetailProject(c *ctx.ServiceCtx, form *forms.DetailProjectForm) (interface{}, e.Error) {
	tx := c.DB().Begin()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()
	//校验用户是否在该项目下有权限
	isExist := IsUserOrgProjectPermission(tx, c.UserId, form.Id, consts.ProjectRoleManager)

	if !isExist {
		return nil, e.New(e.ObjectNotExistsOrNoPerm, http.StatusForbidden, errors.New("not permission"))
	}
	projectUser, err := services.SearchProjectUsers(tx, form.Id)
	if err != nil {
		return nil, e.New(e.DBError, err)
	}
	project, err := services.DetailProject(tx, form.Id)
	if err != nil {
		return nil, e.New(e.DBError, err)
	}

	tplCount, er := services.StatisticalProjectTpl(tx, form.Id)
	if er != nil {
		return nil, e.New(e.DBError, er)
	}
	envResp, er := services.StatisticalProjectEnv(tx, form.Id)
	if er != nil {
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
