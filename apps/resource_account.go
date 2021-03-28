package apps

import (
	"cloudiac/consts/e"
	"cloudiac/libs/ctx"
	"cloudiac/libs/page"
	"cloudiac/models"
	"cloudiac/models/forms"
	"cloudiac/services"
	"fmt"
)

func CreateResourceAccount(c *ctx.ServiceCtx, form *forms.CreateResourceAccountForm) (*models.ResourceAccount, e.Error) {
	c.AddLogField("action", fmt.Sprintf("create resource_account %s", form.Name))

	tx := c.Tx().Debug()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	rsAccount, err := func() (*models.ResourceAccount, e.Error) {
		var (
			rsAccount   *models.ResourceAccount
			err         e.Error
			er          e.Error
		)

		rsAcc := &models.ResourceAccount{
			Name:        form.Name,
			Description: form.Description,
			Params:      []byte(form.Params),
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
				CtServiceId: ctServiceId,
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

type searchResourceAccountResp struct {
	models.ResourceAccount
	CtServiceIds   []string  `json:"ctServiceIds"`
}

func SearchResourceAccount(c *ctx.ServiceCtx, form *forms.SearchResourceAccountForm) (interface{}, e.Error) {
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

func UpdateResourceAccount(c *ctx.ServiceCtx, form *forms.UpdateResourceAccountForm) (rsAccount *models.ResourceAccount, err e.Error) {
	c.AddLogField("action", fmt.Sprintf("update rsAccount %d", form.Id))
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

	if form.HasKey("params") {
		attrs["params"] = []byte(form.Params)
	}

	if form.HasKey("status") {
		attrs["status"] = []byte(form.Status)
	}

	rsAccount, err = services.UpdateResourceAccount(c.DB(), form.Id, attrs)
	if err != nil {
		return nil, err
	}

	if form.HasKey("ctServiceIds") {
		err = services.DeleteCtResourceMap(c.DB(), form.Id)
		if err != nil {
			return nil, err
		}

		for _, ctServiceId := range form.CtServiceIds {
			_, err = services.CreateCtResourceMap(c.DB(), models.CtResourceMap{
				ResourceAccountId: rsAccount.Id,
				CtServiceId: ctServiceId,
			})
			if err != nil {
				return nil, err
			}
		}
	}

	return
}

func DeleteResourceAccount(c *ctx.ServiceCtx, form *forms.DeleteResourceAccountForm) (result interface{}, re e.Error) {
	c.AddLogField("action", fmt.Sprintf("delete ResourceAccount %d for org %d", form.Id, c.OrgId))

	tx := c.Tx()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := services.DeleteResourceAccount(tx, form.Id, c.OrgId); err != nil {
		tx.Rollback()
		return nil, err
	} else if err := tx.Commit(); err != nil {
		return nil, e.New(e.DBError, err)
	}
	c.Logger().Infof("delete ResourceAccount %d", form.Id, " for org %d", c.OrgId, " succeed")

	return
}