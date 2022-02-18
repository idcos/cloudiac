// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package apps

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/libs/page"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
	"cloudiac/utils"
	"encoding/json"
	"fmt"
)

func CreateResourceAccount(c *ctx.ServiceContext, form *forms.CreateResourceAccountForm) (*models.ResourceAccount, e.Error) {
	c.AddLogField("action", fmt.Sprintf("create resource_account %s", form.Name))

	tx := c.Tx()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	rsAccount, err := func() (*models.ResourceAccount, e.Error) {
		var (
			rsAccount *models.ResourceAccount
			err       e.Error
			er        e.Error
		)
		jsons, _ := parseParams(form.Params, map[string]string{})

		rsAcc := &models.ResourceAccount{
			Name:        form.Name,
			Description: form.Description,
			Params:      models.JSON(string(jsons)),
		}
		rsAcc.OrgId = c.OrgId

		rsAccount, err = services.CreateResourceAccount(tx, rsAcc)
		if err != nil {
			return nil, err
		}

		// 绑定资源账号与CT Runner
		for _, ctServiceId := range form.CtServiceIds {
			_, er = services.CreateCtResourceMap(tx, models.CtResourceMap{
				ResourceAccountId: rsAccount.Id,
				CtServiceId:       ctServiceId,
			})
			if er != nil {
				return nil, er
			}
		}

		return rsAccount, nil
	}()
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, e.New(e.DBError, err)
	}

	return rsAccount, nil
}

func parseParams(params []forms.Params, newVars map[string]string) ([]byte, error) {
	for index, v := range params {
		if *v.IsSecret && v.Value != "" {
			encryptedValue, err := utils.AesEncrypt(v.Value)
			params[index].Value = encryptedValue
			if err != nil {
				return nil, err
			}
		}
		if v.Value == "" && *v.IsSecret {
			params[index].Value = newVars[v.Id]
		}
	}
	jsons, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	return jsons, nil
}

type searchResourceAccountResp struct {
	models.ResourceAccount
	CtServiceIds []string `json:"ctServiceIds"`
}

func SearchResourceAccount(c *ctx.ServiceContext, form *forms.SearchResourceAccountForm) (interface{}, e.Error) {
	query := services.QueryResourceAccount(c.DB())
	query = query.Where("org_id = ?", c.OrgId)
	if form.Status != "" {
		query = query.Where("status = ?", form.Status)
	}
	if form.Q != "" {
		qs := "%" + form.Q + "%"
		query = query.Where("name LIKE ? OR description LIKE ? ", qs, qs)
	}

	query = query.Order("created_at DESC")
	p := page.New(form.CurrentPage(), form.PageSize(), query)
	rsAccounts := make([]*searchResourceAccountResp, 0)
	if err := p.Scan(&rsAccounts); err != nil {
		return nil, e.New(e.DBError, err)
	}
	for index, v := range rsAccounts {
		vars := make([]forms.Params, 0)
		//newVars := make([]forms.Params, 0)
		if !v.Params.IsNull() {
			_ = json.Unmarshal(v.Params, &vars)
		}
		newVars := getParams(vars)
		b, _ := json.Marshal(newVars)
		rsAccounts[index].Params = models.JSON(b)
	}

	for _, rsAccount := range rsAccounts {
		ctServiceIds, _ := services.FindCtResourceMap(c.DB(), rsAccount.Id)
		rsAccount.CtServiceIds = ctServiceIds
	}

	return page.PageResp{
		Total:    p.MustTotal(),
		PageSize: p.Size,
		List:     rsAccounts,
	}, nil
}

func chkFormCtServiceIds(db *db.Session, form *forms.UpdateResourceAccountForm, rsAccountId models.Id) e.Error {
	if !form.HasKey("ctServiceIds") {
		return nil
	}

	err := services.DeleteCtResourceMap(db, form.Id)
	if err != nil {
		return err
	}

	for _, ctServiceId := range form.CtServiceIds {
		_, err = services.CreateCtResourceMap(db, models.CtResourceMap{
			ResourceAccountId: rsAccountId,
			CtServiceId:       ctServiceId,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func UpdateResourceAccount(c *ctx.ServiceContext, form *forms.UpdateResourceAccountForm) (rsAccount *models.ResourceAccount, err e.Error) {
	c.AddLogField("action", fmt.Sprintf("update rsAccount %s", form.Id))
	if form.Id == "" {
		return nil, e.New(e.BadRequest, fmt.Errorf("missing 'id'"))
	}

	attrs := models.Attrs{}
	if form.HasKey("name") {
		attrs["name"] = form.Name
	}

	if form.HasKey("description") {
		attrs["description"] = form.Description
	}
	newVars := make(map[string]string, 0)
	vars := make([]forms.Params, 0)
	ra, err := services.GetResourceAccountById(c.DB(), form.Id)
	if err != nil {
		return nil, err
	}
	if !ra.Params.IsNull() {
		_ = json.Unmarshal(ra.Params, &vars)
	}

	for _, v := range vars {
		newVars[v.Id] = v.Value
	}
	if form.HasKey("params") {
		jsons, _ := parseParams(form.Params, newVars)
		attrs["params"] = models.JSON(string(jsons))
	}

	if form.HasKey("status") {
		attrs["status"] = []byte(form.Status)
	}

	rsAccount, err = services.UpdateResourceAccount(c.DB(), form.Id, attrs)
	if err != nil {
		return nil, err
	}

	// check ctServiceIds
	err = chkFormCtServiceIds(c.DB(), form, rsAccount.Id)
	return
}

func DeleteResourceAccount(c *ctx.ServiceContext, form *forms.DeleteResourceAccountForm) (result interface{}, re e.Error) {
	c.AddLogField("action", fmt.Sprintf("delete ResourceAccount %s for org %s", form.Id, c.OrgId))

	tx := c.Tx()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	if err := services.DeleteResourceAccount(tx, form.Id, c.OrgId); err != nil {
		_ = tx.Rollback()
		return nil, err
	} else if err := tx.Commit(); err != nil {
		return nil, e.New(e.DBError, err)
	}
	c.Logger().Infof("delete ResourceAccount %s", form.Id, " for org %s", c.OrgId, " succeed")

	return
}

func getParams(vars []forms.Params) []forms.Params {
	newVars := make([]forms.Params, 0)
	for _, v := range vars {
		if *v.IsSecret {
			newVars = append(newVars, forms.Params{
				Key:      v.Key,
				Value:    "",
				IsSecret: v.IsSecret,
				Id:       v.Id,
			})
		} else {
			newVars = append(newVars, v)
		}
	}
	return newVars
}
