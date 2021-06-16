package apps

import (
	"cloudiac/consts/e"
	"cloudiac/libs/ctx"
	"cloudiac/models"
	"cloudiac/models/forms"
	"cloudiac/services"
	"cloudiac/utils"
	"encoding/json"
)

func SearchMetaTemplate(c *ctx.ServiceCtx, form *forms.SearchTemplateLibraryForm) (interface{}, e.Error) {
	query := services.SearchMetaTemplate(c.DB().Debug())
	rs, _ := getPage(query, form, models.MetaTemplate{})
	return rs, nil
}

func CreateMetaTemplate(c *ctx.ServiceCtx, form *forms.CreateTemplateLibraryForm) (interface{}, e.Error) {
	tx := c.Tx().Debug()
	tplLibVars := make([]forms.Var, 0)
	param := make(map[string]interface{})
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()
	tplLib, err := services.GetMetaTemplateById(tx, form.Id)
	if err != nil {
		return nil, err
	}
	guid := utils.GenGuid("ct")
	template, err := func() (*models.Template, e.Error) {
		var (
			template *models.Template
			err      e.Error
		)

		tplVarsByte, _ := tplLib.Vars.MarshalJSON()
		if !tplLib.Vars.IsNull() {
			_ = json.Unmarshal(tplVarsByte, &tplLibVars)
		}

		for index, v := range tplLibVars {
			if *v.IsSecret {
				encryptedValue, err := utils.AesEncrypt(v.Value)
				if err != nil {
					return nil, nil
				}
				tplLibVars[index].Value = encryptedValue

			}
		}
		jsons, _ := json.Marshal(param)

		template, err = services.CreateTemplate(tx, models.Template{
			OrgId:       c.OrgId,
			Name:        form.Name,
			Guid:        guid,
			Description: tplLib.Description,
			RepoId:      tplLib.RepoId,
			RepoBranch:  tplLib.RepoBranch,
			RepoAddr:    tplLib.RepoAddr,
			SaveState:   tplLib.SaveState,
			Vars:        models.JSON(string(jsons)),
			Varfile:     tplLib.Varfile,
			Timeout:     tplLib.Timeout,
			Creator:     c.UserId,
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
