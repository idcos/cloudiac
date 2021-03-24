package apps

import (
	"cloudiac/consts/e"
	"cloudiac/libs/ctx"
	"cloudiac/libs/page"
	"cloudiac/models"
	"cloudiac/models/forms"
	"cloudiac/services"
	"cloudiac/utils"
	"encoding/json"
	"fmt"
)

func SearchTemplate(c *ctx.ServiceCtx, form *forms.SearchTemplateForm) (interface{}, e.Error) {
	query := services.QueryTemplate(c.DB())
	if form.Status != "" {
		query = query.Where("status = ?", form.Status)
	}
	// if form.Q != "" {
	// 	qs := "%" + form.Q + "%"
	// 	query = query.Where("name LIKE ? OR phone LIKE ? OR email LIKE ? ", qs, qs, qs)
	// }

	query = query.Order("created_at DESC")
	p := page.New(form.CurrentPage(), form.PageSize(), query)
	templates := make([]*models.Template, 0)
	if err := p.Scan(&templates); err != nil {
		return nil, e.New(e.DBError, err)
	}

	return page.PageResp{
		Total:    p.MustTotal(),
		PageSize: p.Size,
		List:     templates,
	}, nil
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

	guid := utils.GenGuid("ct")
	template, err := func() (*models.Template, e.Error) {
		var (
			template *models.Template
			err      e.Error
		)

		vars := form.Vars
		for index, v := range vars {
			if v.IsSecret == true {
				encryptedValue, err := utils.AesEncrypt(v.Value)
				vars[index].Value = encryptedValue
				if err != nil {
					return nil, nil
				}
			}
		}
		jsons, _ := json.Marshal(vars)

		template, err = services.CreateTemplate(tx, models.Template{
			Name:        form.Name,
			Guid:        guid,
			Description: form.Description,
			RepoId:      form.RepoId,
			RepoBranch:  form.RepoBranch,
			RepoAddr:    form.RepoAddr,
			SaveState:   form.SaveState,
			Vars:        models.JSON(string(jsons)),
			Varfile:     form.Varfile,
			Extra:       form.Extra,
			Timeout:     form.Timeout,
		})
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

func UpdateTemplate(c *ctx.ServiceCtx, form *forms.UpdateTemplateForm) (user *models.Template, err e.Error) {
	c.AddLogField("action", fmt.Sprintf("update template %d", form.Id))
	if form.Id == 0 {
		return nil, e.New(e.BadRequest, fmt.Errorf("missing 'id'"))
	}

	attrs := models.Attrs{}
	if form.HasKey("name") {
		attrs["name"] = form.Name
	}

	if form.HasKey("description") {
		attrs["description"] = form.Description
	}

	if form.HasKey("vars") {
		vars := form.Vars
		for index, v := range vars {
			if v.IsSecret == true {
				encryptedValue, err := utils.AesEncrypt(v.Value)
				vars[index].Value = encryptedValue
				if err != nil {
					return nil, nil
				}
			}
		}
		jsons, _ := json.Marshal(vars)
		attrs["vars"] = jsons
	}

	if form.HasKey("varfile") {
		attrs["varfile"] = form.Varfile
	}

	if form.HasKey("extra") {
		attrs["extra"] = form.Extra
	}

	if form.HasKey("timeout") {
		attrs["timeout"] = form.Timeout
	}

	user, err = services.UpdateTemplate(c.DB(), form.Id, attrs)
	return
}
