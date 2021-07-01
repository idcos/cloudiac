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
	if err != nil {
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
	rs, _ := getPage(query, form, &models.Organization{})
	return rs, nil
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
func UpdateOrganization(c *ctx.ServiceCtx, orgId models.Id, form *forms.UpdateOrganizationForm) (org *models.Organization, err e.Error) {
	c.AddLogField("action", fmt.Sprintf("update org %d", orgId))
	if orgId == "" {
		return nil, e.New(e.BadRequest, fmt.Errorf("missing 'id'"))
	}

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

	org, err = services.UpdateOrganization(c.DB(), orgId, attrs)
	return
}

func ChangeOrgStatus(c *ctx.ServiceCtx, orgId models.Id, form *forms.DisableOrganizationForm) (org *models.Organization, err e.Error) {
	c.AddLogField("action", fmt.Sprintf("change org status %s", orgId))

	if form.Status != models.OrgEnable && form.Status != models.OrgDisable {
		return nil, e.New(e.OrganizationInvalidStatus, http.StatusBadRequest)
	}

	org, err = services.GetOrganizationById(c.DB(), orgId)
	if err != nil {
		if err.Code() == e.OrganizationNotExists {
			return nil, e.New(err.Code(), err, http.StatusBadRequest)
		}
		return nil, err
	}

	if org.Status == form.Status {
		return org, nil
	}

	org, err = services.UpdateOrganization(c.DB(), orgId, models.Attrs{"status": form.Status})
	if err != nil {
		return nil, e.New(e.DBError, err)
	}

	return org, nil
}

type organizationDetailResp struct {
	models.Organization
	Creator string `example:"超级管理员"`
}

func ModelIdInArray(v models.Id, arr ...models.Id) bool {
	for i := range arr {
		if arr[i] == v {
			return true
		}
	}
	return false
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
func OrganizationDetail(c *ctx.ServiceCtx, form forms.DetailOrganizationForm) (resp interface{}, er e.Error) {
	orgIds, err := services.GetOrgIdsByUser(c.DB(), c.UserId)
	if err != nil {
		return nil, e.New(e.DBError, err)
	}

	if ModelIdInArray(form.Id, orgIds...) == false && c.IsSuperAdmin == false {
		return nil, e.New(e.OrganizationNotExists, http.StatusNotFound)
	}
	org, err2 := services.GetOrganizationById(c.DB(), form.Id)
	if err2 != nil {
		if err2.Code() == e.OrganizationNotExists {
			return nil, e.New(e.OrganizationNotExists, err2, http.StatusNotFound)
		}
		c.Logger().Error("db error while get org detail, err %s", err)
		return nil, e.New(e.DBError, err)
	}
	user, err3 := services.GetUserById(c.DB(), org.CreatorId)
	if err3 != nil {
		if err3.Code() == e.UserNotExists {
			// 报 500 错误，正常情况用户不应该找不到
			return nil, e.New(e.UserNotExists, err3, http.StatusInternalServerError)
		}
		c.Logger().Error("db error while get detail, err %s", err)
		return nil, e.New(e.DBError, err)
	}
	var o = organizationDetailResp{
		Organization: *org,
		Creator:      user.Name,
	}

	return o, nil
}
