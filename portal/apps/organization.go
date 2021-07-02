package apps

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
	"fmt"
	"net/http"
)

// CreateOrganization 创建组织
// @Tags 组织
// @Description 创建组织接口
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Param Authorization header string true "认证令牌"
// @Param form formData forms.CreateOrganizationForm true "parameter"
// @router /api/v1/orgs [post]
// @Success 200 {object} ctx.JSONResult{result=models.Organization}
func CreateOrganization(c *ctx.ServiceCtx, form *forms.CreateOrganizationForm) (*models.Organization, e.Error) {
	c.AddLogField("action", fmt.Sprintf("create org %s", form.Name))

	org, err := services.CreateOrganization(c.DB(), models.Organization{
		Name:        form.Name,
		CreatorId:   c.UserId,
		Description: form.Description,
	})
	if err != nil && err.Code() == e.OrganizationAlreadyExists {
		return nil, e.New(err.Code(), err, http.StatusBadRequest)
	} else if err != nil {
		c.Logger().Errorf("error creating org, err %s", err)
		return nil, e.AutoNew(err, e.DBError)
	}
	return org, nil
}

// SearchOrganization 组织查询
// @Tags 组织
// @Description 组织查询接口
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Authorization header string true "认证令牌"
// @Param form query forms.SearchOrganizationForm true "parameter"
// @router /api/v1/orgs [get]
// @Success 200 {object} ctx.JSONResult{result=page.PageResp{list=[]models.Organization}}
func SearchOrganization(c *ctx.ServiceCtx, form *forms.SearchOrganizationForm) (interface{}, e.Error) {
	query := services.QueryOrganization(c.DB())
	if c.IsSuperAdmin == true {
		if form.Status != "" {
			query = query.Where("status = ?", form.Status)
		}
	} else {
		query = query.Where("status = 'enable'")
		orgIds, er := services.GetOrgIdsByUser(c.DB(), c.UserId)
		if er != nil {
			c.Logger().Errorf("error get org id by user, err %s", er)
			return nil, e.New(e.DBError, er)
		}
		query = query.Where("id in (?)", orgIds)
	}

	if form.Q != "" {
		qs := "%" + form.Q + "%"
		query = query.WhereLike("name", qs)
	}

	// 默认按创建时间逆序排序
	if form.SortField() == "" {
		query = query.Order("created_at DESC")
	}
	rs, err := getPage(query, form, &models.Organization{})
	if err != nil {
		c.Logger().Errorf("error get page, err %s", err)
	}
	return rs, err
}

// UpdateOrganization 组织编辑
// @Tags 组织
// @Description 组织信息编辑接口
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Param Authorization header string true "认证令牌"
// @Param orgId path string true "组织ID"
// @Param form formData forms.UpdateOrganizationForm true "parameter"
// @router /api/v1/orgs/{orgId} [put]
// @Success 200 {object} ctx.JSONResult{result=models.Organization}
func UpdateOrganization(c *ctx.ServiceCtx, orgId models.Id, form *forms.UpdateOrganizationForm) (*models.Organization, e.Error) {
	c.AddLogField("action", fmt.Sprintf("update org %s", orgId))

	attrs := models.Attrs{}
	if form.HasKey("name") {
		attrs["name"] = form.Name
	}

	if form.HasKey("description") {
		attrs["description"] = form.Description
	}

	if form.HasKey("runnerId") {
		attrs["runner_id"] = form.RunnerId
	}

	// 变更组织状态
	if form.HasKey("status") {
		if _, err := ChangeOrgStatus(c, orgId, &forms.DisableOrganizationForm{Status: form.Status}); err != nil {
			return nil, err
		}
	}

	org, err := services.UpdateOrganization(c.DB(), orgId, attrs)
	if err != nil && err.Code() == e.OrganizationAliasDuplicate {
		return nil, e.New(err.Code(), err, http.StatusBadRequest)
	} else if err != nil {
		c.Logger().Errorf("error update org, err %s", err)
		return nil, err
	}
	return org, nil
}

//ChangeOrgStatus 修改组织启用/禁用状态
func ChangeOrgStatus(c *ctx.ServiceCtx, orgId models.Id, form *forms.DisableOrganizationForm) (*models.Organization, e.Error) {
	c.AddLogField("action", fmt.Sprintf("change org status %s", orgId))

	if form.Status != models.OrgEnable && form.Status != models.OrgDisable {
		return nil, e.New(e.OrganizationInvalidStatus, http.StatusBadRequest)
	}

	org, err := services.GetOrganizationById(c.DB(), orgId)
	if err != nil && err.Code() == e.OrganizationNotExists {
		return nil, e.New(err.Code(), err, http.StatusBadRequest)
	} else if err != nil {
		c.Logger().Errorf("error get org by id, err %s", err)
		return nil, err
	}

	if org.Status == form.Status {
		return org, nil
	}

	org, err = services.UpdateOrganization(c.DB(), orgId, models.Attrs{"status": form.Status})
	if err != nil && err.Code() == e.OrganizationAliasDuplicate {
		return nil, e.New(err.Code(), err, http.StatusBadRequest)
	} else if err != nil {
		c.Logger().Errorf("error update org, err %s", err)
		return nil, e.New(e.DBError, err)
	}

	return org, nil
}

type organizationDetailResp struct {
	models.Organization
	Creator string `example:"超级管理员"`
}

// OrganizationDetail 组织信息详情
// @Tags 组织
// @Description 组织信息详情查询接口
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Authorization header string true "认证令牌"
// @Param orgId path string true "组织ID"
// @router /api/v1/orgs/{orgId} [get]
// @Success 200 {object} ctx.JSONResult{result=organizationDetailResp}
func OrganizationDetail(c *ctx.ServiceCtx, form forms.DetailOrganizationForm) (*organizationDetailResp, e.Error) {
	orgIds, er := services.GetOrgIdsByUser(c.DB(), c.UserId)
	if er != nil {
		c.Logger().Errorf("error get org id by user, err %s", er)
		return nil, e.New(e.DBError, er)
	}
	if form.Id.InArray(orgIds...) == false && c.IsSuperAdmin == false {
		// 请求了一个不存在的 org，因为 org id 是在 path 传入，这里我们返回 404
		return nil, e.New(e.OrganizationNotExists, http.StatusNotFound)
	}

	var (
		org  *models.Organization
		user *models.User
		err  e.Error
	)
	org, err = services.GetOrganizationById(c.DB(), form.Id)
	if err != nil && err.Code() == e.OrganizationNotExists {
		return nil, e.New(e.OrganizationNotExists, err, http.StatusNotFound)
	} else if err != nil {
		c.Logger().Errorf("error get org by id, err %s", err)
		return nil, e.New(e.DBError, err)
	}
	user, err = services.GetUserById(c.DB(), org.CreatorId)
	if err != nil && err.Code() == e.UserNotExists {
		// 报 500 错误，正常情况用户不应该找不到，除非被意外删除
		return nil, e.New(e.UserNotExists, err, http.StatusInternalServerError)
	} else if err != nil {
		c.Logger().Errorf("error get user by id, err %s", err)
		return nil, e.New(e.DBError, err)
	}

	var o = organizationDetailResp{
		Organization: *org,
		Creator:      user.Name,
	}

	return &o, nil
}

// DeleteOrganization 删除组织
// @Tags 组织
// @Description 删除组织接口
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Param Authorization header string true "认证令牌"
// @Param orgId path string true "组织ID"
// @Param form formData forms.DeleteOrganizationForm true "parameter"
// @router /api/v1/orgs/{orgId} [delete]
// @Success 501 {object} ctx.JSONResult
func DeleteOrganization(c *ctx.ServiceCtx, orgId models.Id, form *forms.DeleteOrganizationForm) (org *models.Organization, err e.Error) {
	c.AddLogField("action", fmt.Sprintf("delete org %s", orgId))
	return nil, e.New(e.BadRequest, http.StatusNotImplemented)
}
