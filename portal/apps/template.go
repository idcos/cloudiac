package apps

import (
	"cloudiac/portal/libs/page"
	"fmt"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
	"net/http"
)

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
			tpl      models.Template
		)
		template, err = services.CreateTemplate(tx, tpl)
		if err != nil {
			return nil, err
		}

		return template, nil
	}()

	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, e.New(e.DBError, err)
	}

	return template, nil
}

func UpdateTemplate(c *ctx.ServiceCtx, form *forms.UpdateTemplateForm) (*models.Template, e.Error) {
	c.AddLogField("action", fmt.Sprintf("update template %d", form.Id))
	// TODO 变量相关的先不管
	//vars := make([]forms.Var, 0)
	//newVars := make(map[string]string, 0)
	//for _, v := range vars {
	//	newVars[v.Id] = v.Value
	//}
	//
	//if form.HasKey("vars") {
	//	vars := form.Vars
	//	for index, v := range vars {
	//		if *v.IsSecret && v.Value != "" {
	//			encryptedValue, err := utils.AesEncrypt(v.Value)
	//			vars[index].Value = encryptedValue
	//			if err != nil {
	//				return nil, nil
	//			}
	//		}
	//		if v.Value == "" && *v.IsSecret {
	//			vars[index].Value = newVars[v.Id]
	//		}
	//	}
	//	jsons, _ := json.Marshal(vars)
	//	attrs["vars"] = jsons
	//}
	//
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


	if form.HasKey("varfile") {
		attrs["varfile"] = form.Varfile
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
	if form.HasKey("runnerId") {
		attrs["runnerId"] = form.RunnerId
	}

	return services.UpdateTemplate(c.DB().Debug(), form.Id, attrs)
}

func DelateTemplate(c *ctx.ServiceCtx, form *forms.DeleteTemplateForm) (interface{}, e.Error){
	c.AddLogField("action", fmt.Sprintf("delete template %d", form.Id))

	// TODO 判断云模版是否属于该组织
	// TODO 判断云模版是否活跃

	tx := c.Tx().Debug()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()
	// 根据ID 查询云模版是否存在
	tpl, err := services.GetTemplateById(c.DB(), form.Id)
	if err != nil && err.Code() == e.UserNotExists {
		return nil, e.New(err.Code(), err, http.StatusNotFound)
	} else if err != nil {
		c.Logger().Errorf("error get template by id, err %s", err)
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	}
	// 根据ID 删除云模版
	if err := services.DeleteTemplate(tx, tpl.Id); err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("error commit del template, err %s", err)
		return nil, e.New(e.DBError, err)
	}

	return nil, nil

}

func TemplateDetail(c *ctx.ServiceCtx, form *forms.DetailTemplateForm) (*models.Template, e.Error) {
	tpl, err := services.GetTemplateById(c.DB(), form.Id)
	if err != nil && err.Code() == e.TaskNotExists {
		return nil, e.New(err.Code(), err, http.StatusNotFound)
	} else if err != nil {
		c.Logger().Errorf("error get template by id, err %s", err)
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	}
	return tpl, nil

}



type SearchTemplateResp struct {
	Id            uint      `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	ActiveEnvironment int   `json:"activeEnvironment"`
	VcsType		  string    `json:"vcsType"`
	RepoRevision  string  	`json:"repoRevision"`
	UserName 	  string    `json:"userName"`
	CreateTime    string	`json:"createTime"`

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


