// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package apps

import (
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/models/resps"
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

	query = query.
		Joins("LEFT JOIN iac_user ON iac_user.id = iac_key.creator_id").
		Select("iac_key.*, iac_user.name AS creator")
	rs, err := getPage(query, form, resps.KeyResp{})
	if err != nil {
		c.Logger().Errorf("error search key, err %s", err)
		return nil, err
	}

	return rs, nil
}

// CreateKey 创建密钥
func CreateKey(c *ctx.ServiceContext, form *forms.CreateKeyForm) (interface{}, e.Error) {
	c.AddLogField("action", fmt.Sprintf("create key %s", form.Name))

	if !c.IsSuperAdmin && !services.UserHasOrgRole(c.UserId, c.OrgId, consts.OrgRoleAdmin) {
		projectRole, err := services.GetUserHighestProjectRole(c.DB(), c.OrgId, c.UserId)
		if err != nil {
			return nil, err
		} else if projectRole == consts.ProjectRoleGuest {
			// 除了组织管理员，项目的非 guest 角色也可以添加密钥对，但只能管理自己添加的密钥对
			return nil, e.New(e.PermissionDeny, http.StatusForbidden)
		}
	}

	encrypted, er := utils.AesEncrypt(form.Key)
	if er != nil {
		return nil, e.New(e.InternalError, fmt.Errorf("error encrypt key"), http.StatusInternalServerError)
	}
	key, err := services.CreateKey(c.DB(), models.Key{
		OrgId:     c.OrgId,
		Name:      form.Name,
		Content:   models.Text(encrypted),
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

	key = &models.Key{}
	if err := query.Find(key, form.Id); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}
	if ok, err := hasKeyDataPerm(c, key); err != nil {
		return nil, e.AutoNew(err, e.InternalError)
	} else if ok {
		return nil, e.New(e.PermissionDeny, http.StatusOK)
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

	key := models.Key{}
	if err := query.Find(&key, form.Id); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}

	if ok, err := hasKeyDataPerm(c, &key); err != nil {
		return nil, e.AutoNew(err, e.InternalError)
	} else if !ok {
		return nil, e.New(e.PermissionDeny, http.StatusOK)
	}

	if err := services.DeleteKey(query, form.Id); err != nil {
		return nil, err
	}
	return
}

func DetailKey(c *ctx.ServiceContext, form *forms.DetailKeyForm) (result interface{}, re e.Error) {
	query := services.QueryKey(services.QueryWithOrgId(c.DB(), c.OrgId))
	key, err := services.GetKeyById(query, form.Id, false)
	if err != nil {
		if err.Code() == e.KeyNotExist {
			return nil, e.New(err.Code(), err, http.StatusBadRequest)
		}
		c.Logger().Errorf("error get key by id, err %s", err)
		return nil, err
	}

	user, err := services.GetUserById(c.DB(), key.CreatorId)
	if err != nil {
		return nil, err
	}

	return resps.KeyResp{
		Key:     *key,
		Creator: user.Name,
	}, nil
}

func hasKeyDataPerm(c *ctx.ServiceContext, key *models.Key) (bool, error) {
	if c.IsSuperAdmin {
		return true, nil
	}
	if c.ProjectId != "" && services.UserHasOrgRole(c.UserId, c.OrgId, consts.OrgRoleAdmin) {
		return true, nil
	}

	if c.UserId == key.CreatorId {
		role, err := services.GetUserHighestProjectRole(c.DB(), c.OrgId, c.UserId)
		if err != nil {
			return false, e.AutoNew(err, e.DBError)
		}
		if role != consts.ProjectRoleGuest {
			return true, nil
		}
	}
	return false, nil
}
