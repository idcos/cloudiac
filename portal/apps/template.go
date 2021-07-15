package apps

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/libs/page"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
	"fmt"
	"net/http"
)

type SearchTemplateResp struct {
	Id                uint   `json:"id"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	ActiveEnvironment int    `json:"activeEnvironment"`
	VcsType           string `json:"vcsType"`
	RepoRevision      string `json:"repoRevision"`
	UserName          string `json:"userName"`
	CreateTime        string `json:"createTime"`
}

func CreateTemplate(c *ctx.ServiceCtx, form *forms.CreateTemplateForm) (*models.Template, e.Error) {
	c.AddLogField("action", fmt.Sprintf("create template %s", form.Name))

	tx := c.Tx().Debug()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	template, err := func() (*models.Template, e.Error) {
		var (
			template *models.Template
			err      e.Error
		)
		tpl := models.Template{
			Name:         form.Name,
			TplType:      form.TplType,
			OrgId:        c.OrgId,
			Description:  form.Description,
			VcsId:        form.VcsId,
			RepoId:       form.RepoId,
			RepoAddr:     form.RepoAddr,
			RepoRevision: form.RepoRevision,
			CreatorId:    c.UserId,
			Workdir:      form.Workdir,
			Playbook:     form.Playbook,
			PlayVarsFile: form.PlayVarsFile,
			TfVarsFile:   form.TfVarsFile,
		}
		template, err = services.CreateTemplate(tx, tpl)
		if err != nil {
			return nil, err
		}
		// TODO 创建 项目模板一对多关联
		// TODO 创建 变量

		return template, nil
	}()

	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("error commit create template, err %s", err)
		return nil, e.New(e.DBError, err)
	}

	return template, nil
}

func UpdateTemplate(c *ctx.ServiceCtx, form *forms.UpdateTemplateForm) (*models.Template, e.Error) {
	c.AddLogField("action", fmt.Sprintf("update template %d", form.Id))

	tpl, err := services.GetTemplate(c.DB(), form.Id)
	if err != nil {
		return nil, e.New(e.DBError, err, e.TemplateNotExists)
	}
	// 根据云模板ID, 组织ID查询该云模板是否属于该组织
	if tpl.OrgId != c.OrgId {
		return nil, e.New(e.TemplateNotExists, http.StatusForbidden, fmt.Errorf("The organization does not have permission to delete the current template"))
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
	if form.HasKey("extra") {
		attrs["extra"] = form.Extra
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

	return services.UpdateTemplate(c.DB(), form.Id, attrs)
}

func DelateTemplate(c *ctx.ServiceCtx, form *forms.DeleteTemplateForm) (interface{}, e.Error) {
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

func TemplateDetail(c *ctx.ServiceCtx, form *forms.DetailTemplateForm) (*models.Template, e.Error) {
	tpl, err := services.GetTemplateById(c.DB(), form.Id)
	if err != nil && err.Code() == e.TemplateNotExists {
		return nil, e.New(err.Code(), err, http.StatusNotFound)
	} else if err != nil {
		c.Logger().Errorf("error get template by id, err %s", err)
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	}
	return tpl, nil

}

func SearchTemplate(c *ctx.ServiceCtx, form *forms.SearchTemplateForm) (tpl interface{}, err e.Error) {
	tplIdList := make([]string, 0)
	if c.ProjectId != "" {
		tplIdList, err = services.QueryTplByProjectId(c.DB(), c.ProjectId)
		if err != nil {
			return nil, err
		}
	}
	query, _ := services.QueryTemplate(c.DB().Debug(), form.Q, c.OrgId, tplIdList)
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
