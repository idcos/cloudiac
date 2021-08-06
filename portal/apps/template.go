// Copyright 2021 CloudJ Company Limited. All rights reserved.

package apps

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/libs/page"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
	"cloudiac/portal/services/vcsrv"
	"fmt"
	"net/http"
)

type SearchTemplateResp struct {
	CreatedAt         models.Time `json:"createdAt"` // 创建时间
	UpdatedAt         models.Time `json:"updatedAt"` // 更新时间
	Id                models.Id   `json:"id"`
	Name              string      `json:"name"`
	Description       string      `json:"description"`
	ActiveEnvironment int         `json:"activeEnvironment"`
	RepoRevision      string      `json:"repoRevision"`
	Creator           string      `json:"creator"`
	RepoId            string      `json:"repoId"`
	VcsId             string      `json:"vcsId"`
	RepoAddr          string      `json:"repoAddr"`
}

func getRepoAddr(vcsId models.Id, query *db.Session, repoId string) (string, error) {
	vcs, err := services.QueryVcsByVcsId(vcsId, query)
	if err != nil {
		return "", err
	}
	vcsIface, er := vcsrv.GetVcsInstance(vcs)
	if er != nil {
		return "", er
	}
	repo, er := vcsIface.GetRepo(repoId)
	if er != nil {
		return "", er
	}
	repoAddr, er := vcsrv.GetRepoAddress(repo)
	if er != nil {
		return "", er
	}
	return repoAddr, nil
}

func CreateTemplate(c *ctx.ServiceContext, form *forms.CreateTemplateForm) (*models.Template, e.Error) {
	c.AddLogField("action", fmt.Sprintf("create template %s", form.Name))

	repoAddr, er := getRepoAddr(form.VcsId, c.DB(), form.RepoId)
	if er != nil {
		return nil, e.New(e.DBError, fmt.Errorf("get repo failed: %v", er))
	}

	tx := c.Tx().Debug()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()
	template, err := services.CreateTemplate(tx, models.Template{
		Name:         form.Name,
		OrgId:        c.OrgId,
		Description:  form.Description,
		VcsId:        form.VcsId,
		RepoId:       form.RepoId,
		RepoAddr:     repoAddr,
		RepoRevision: form.RepoRevision,
		CreatorId:    c.UserId,
		Workdir:      form.Workdir,
		Playbook:     form.Playbook,
		PlayVarsFile: form.PlayVarsFile,
		TfVarsFile:   form.TfVarsFile,
	})

	if err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("error create template, err %s", err)
		if err.Code() == e.TemplateAlreadyExists {
			return nil, e.New(err.Code(), err.Err(), http.StatusBadRequest)
		}
		return nil, err
	}

	// 创建模板与项目的关系
	if err := services.CreateTemplateProject(tx, form.ProjectId, template.Id); err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	// 创建变量
	if err := services.OperationVariables(tx, c.OrgId, c.ProjectId,
		template.Id, "", form.Variables, form.DeleteVariablesId); err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("error operation variables, err %s", err)
		return nil, e.New(e.DBError, err)
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("error commit create template, err %s", err)
		return nil, e.New(e.DBError, err)
	}

	return template, nil
}

func UpdateTemplate(c *ctx.ServiceContext, form *forms.UpdateTemplateForm) (*models.Template, e.Error) {
	c.AddLogField("action", fmt.Sprintf("update template %s", form.Id))

	tpl, err := services.GetTemplateById(c.DB(), form.Id)
	if err != nil {
		return nil, e.New(e.TemplateNotExists, err, http.StatusBadRequest)
	}

	// 根据云模板ID, 组织ID查询该云模板是否属于该组织
	if tpl.OrgId != c.OrgId {
		return nil, e.New(e.TemplateNotExists, http.StatusForbidden, fmt.Errorf("the organization does not have permission to delete the current template"))
	}
	attrs := models.Attrs{}
	if form.HasKey("name") {
		attrs["name"] = form.Name
	}

	if form.HasKey("description") {
		attrs["description"] = form.Description
	}
	if form.HasKey("playbook") {
		attrs["playbook"] = form.Playbook
	}
	if form.HasKey("status") {
		attrs["status"] = form.Status
	}
	if form.HasKey("workdir") {
		attrs["workdir"] = form.Workdir
	}
	if form.HasKey("tfVarsFile") {
		attrs["tfVarsFile"] = form.TfVarsFile
	}
	if form.HasKey("playVarsFile") {
		attrs["playVarsFile"] = form.PlayVarsFile
	}
	if form.HasKey("repoRevision") {
		attrs["repoRevision"] = form.RepoRevision
	}
	if form.HasKey("vcsId") && form.HasKey("repoId") {
		repoAddr, er := getRepoAddr(form.VcsId, c.DB(), form.RepoId)
		if er != nil {
			return nil, e.New(e.DBError, fmt.Errorf("get repo failed: %v", er))
		}
		attrs["repoAddr"] = repoAddr
		attrs["vcsId"] = form.VcsId
		attrs["repoId"] = form.RepoId
	}
	tx := c.Tx()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()
	if tpl, err = services.UpdateTemplate(tx, form.Id, attrs); err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	if form.HasKey("projectId") {
		if err := services.DeleteTemplateProject(tx, form.Id); err != nil {
			_ = tx.Rollback()
			return nil, err
		}
		if err := services.CreateTemplateProject(tx, form.ProjectId, form.Id); err != nil {
			_ = tx.Rollback()
			return nil, err
		}
	}
	if form.HasKey("variables") || form.HasKey("deleteVariablesId") {
		if err := services.OperationVariables(tx, c.OrgId, c.ProjectId,
			form.Id, "", form.Variables, form.DeleteVariablesId); err != nil {
			_ = tx.Rollback()
			c.Logger().Errorf("error operation variables, err %s", err)
			return nil, e.New(e.DBError, err)
		}
	}
	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("error commit update template, err %s", err)
		return nil, e.New(e.DBError, err)
	}
	return tpl, err
}

func DeleteTemplate(c *ctx.ServiceContext, form *forms.DeleteTemplateForm) (interface{}, e.Error) {
	c.AddLogField("action", fmt.Sprintf("delete template %s", form.Id))
	tx := c.Tx()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()
	// 根据ID 查询云模板是否存在
	tpl, err := services.GetTemplateById(tx, form.Id)
	if err != nil && err.Code() == e.TemplateNotExists {
		return nil, e.New(err.Code(), err, http.StatusNotFound)
	} else if err != nil {
		c.Logger().Errorf("error get template by id, err %v", err)
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	}
	// 根据云模板ID, 组织ID查询该云模板是否属于该组织
	if tpl.OrgId != c.OrgId {
		return nil, e.New(e.TemplateNotExists, http.StatusForbidden, fmt.Errorf("The organization does not have permission to delete the current template"))
	}
	// 查询活跃环境
	envList, er := services.GetEnvByTplId(tx, form.Id)
	if er != nil {
		return nil, e.AutoNew(er, e.DBError)
	}
	for _, v := range envList {
		if v.Status != "inactive" {
			c.Logger().Error("error delete template by id,because the template also has an active environment")
			return nil, e.New(e.TemplateActiveEnvExists, http.StatusForbidden, fmt.Errorf("The cloud template cannot be deleted because there is an active environment"))
		}
	}
	// 根据ID 删除云模板
	if err := services.DeleteTemplate(tx, tpl.Id); err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("error commit del template, err %s", err)
		return nil, e.New(e.DBError, err)
	}
	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("error commit del template, err %s", err)
		return nil, e.New(e.DBError, err)
	}

	return nil, nil

}

type TemplateDetailResp struct {
	*models.Template
	Variables   []models.Variable `json:"variables"`
	ProjectList []models.Id       `json:"projectId"`
}

func TemplateDetail(c *ctx.ServiceContext, form *forms.DetailTemplateForm) (*TemplateDetailResp, e.Error) {
	tpl, err := services.GetTemplateById(c.DB(), form.Id)
	if err != nil && err.Code() == e.TemplateNotExists {
		return nil, e.New(err.Code(), err, http.StatusNotFound)
	} else if err != nil {
		c.Logger().Errorf("error get template by id, err %s", err)
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	}
	project_ids, err := services.QuertProjectByTplId(c.DB(), form.Id)
	if err != nil {
		return nil, e.New(e.DBError, err)
	}
	varialbeList, err := services.SearchVariableByTemplateId(c.DB(), form.Id)
	if err != nil {
		return nil, e.New(e.DBError, err)
	}
	tplDetail := &TemplateDetailResp{
		Template:    tpl,
		Variables:   varialbeList,
		ProjectList: project_ids,
	}
	return tplDetail, nil

}

func SearchTemplate(c *ctx.ServiceContext, form *forms.SearchTemplateForm) (tpl interface{}, err e.Error) {
	tplIdList := make([]models.Id, 0)
	if c.ProjectId != "" {
		// TODO 是否校验project_id 在不在这个组织里面？
		tplIdList, err = services.QueryTplByProjectId(c.DB(), c.ProjectId)
		if err != nil {
			return nil, err
		}

		if len(tplIdList) == 0 {
			return getEmptyListResult(form)
		}
	}
	query := services.QueryTemplateByOrgId(c.DB(), form.Q, c.OrgId, tplIdList)
	p := page.New(form.CurrentPage(), form.PageSize(), query)
	templates := make([]*SearchTemplateResp, 0)
	if err := p.Scan(&templates); err != nil {
		return nil, e.New(e.DBError, err)
	}

	return page.PageResp{
		Total:    p.MustTotal(),
		PageSize: p.Size,
		List:     templates,
	}, nil
}
