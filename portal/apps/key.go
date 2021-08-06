// Copyright 2021 CloudJ Company Limited. All rights reserved.

package apps

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
	"cloudiac/utils"
	"fmt"
	"net/http"
)

// SearchKey 密钥列表查询
func SearchKey(c *ctx.ServiceContext, form *forms.SearchKeyForm) (interface{}, e.Error) {
	query := services.QueryKey(services.QueryWithOrgId(c.DB(), c.OrgId))
	if form.Q != "" {
		qs := "%" + form.Q + "%"
		query = query.Where("name LIKE ?", qs)
	}

	if form.SortField() == "" {
		query = query.Order("created_at DESC")
	}

	rs, err := getPage(query, form, models.Key{})
	if err != nil {
		c.Logger().Errorf("error search key, err %s", err)
		return nil, err
	}

	return rs, nil
}

// CreateKey 创建密钥
func CreateKey(c *ctx.ServiceContext, form *forms.CreateKeyForm) (interface{}, e.Error) {
	c.AddLogField("action", fmt.Sprintf("create key %s", form.Name))

	encrypted, er := utils.AesEncrypt(form.Key)
	if er != nil {
		return nil, e.New(e.InternalError, fmt.Errorf("error encrypt key"), http.StatusInternalServerError)
	}
	key, err := services.CreateKey(c.DB(), models.Key{
		OrgId:     c.OrgId,
		Name:      form.Name,
		Content:   encrypted,
		CreatorId: c.UserId,
	})
	if err != nil && err.Code() == e.KeyAlreadyExists {
		return nil, e.New(err.Code(), err, http.StatusBadRequest)
	} else if err != nil {
		c.Logger().Errorf("error creating key, err %s", err)
		return nil, e.AutoNew(err, e.DBError)
	}
	return key, nil
}

func UpdateKey(c *ctx.ServiceContext, form *forms.UpdateKeyForm) (key *models.Key, err e.Error) {
	c.AddLogField("action", fmt.Sprintf("update key %s", form.Id))
	if form.Id == "" {
		return nil, e.New(e.BadRequest, fmt.Errorf("missing 'id'"))
	}
	query := services.QueryKey(services.QueryWithOrgId(c.DB(), c.OrgId))

	attrs := models.Attrs{}
	if form.HasKey("name") {
		attrs["name"] = form.Name
	}

	key, err = services.UpdateKey(query, form.Id, attrs)
	if err != nil && (err.Code() == e.KeyAliasDuplicate || err.Code() == e.KeyNotExist) {
		return nil, e.New(err.Code(), err, http.StatusBadRequest)
	} else if err != nil {
		c.Logger().Errorf("error update key, err %s", err)
		return nil, err
	}

	return
}

func DeleteKey(c *ctx.ServiceContext, form *forms.DeleteKeyForm) (result interface{}, re e.Error) {
	c.AddLogField("action", fmt.Sprintf("delete key %s", form.Id))
	query := services.QueryKey(services.QueryWithOrgId(c.DB(), c.OrgId))
	if err := services.DeleteKey(query, form.Id); err != nil {
		return nil, err
	}

	return
}

func DetailKey(c *ctx.ServiceContext, form *forms.DetailKeyForm) (result interface{}, re e.Error) {
	query := services.QueryKey(services.QueryWithOrgId(c.DB(), c.OrgId))
	if key, err := services.GetKeyById(query, form.Id, false); err != nil {
		if err.Code() == e.KeyNotExist {
			return nil, e.New(err.Code(), err, http.StatusBadRequest)
		}
		c.Logger().Errorf("error get key by id, err %s", err)
		return nil, err
	} else {
		return key, nil
	}
}
