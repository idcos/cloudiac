package apps

import (
	"cloudiac/consts/e"
	"cloudiac/libs/ctx"
	"cloudiac/models"
	"cloudiac/models/forms"
	"cloudiac/services"
	"cloudiac/utils"
	"fmt"
	"net/http"
)


func CreateOrganization(c *ctx.ServiceCtx, form *forms.CreateOrganizationForm) (*models.Organization, e.Error) {
	c.AddLogField("action", fmt.Sprintf("create org %s", form.Name))

	// todo: 验证vcs_provider信息是否有效，一期固定使用内部gitlab，暂不实现

	guid := utils.GenGuid("org")
	org, err := services.CreateOrganization(c.DB(), models.Organization{
		Name:        form.Name,
		Guid:        guid,
		Description: form.Description,
		Creator:     c.UserId,
		VcsType:     form.VcsType,
		VcsVersion:  form.VcsVersion,
		VcsAuthInfo: form.VcsAuthInfo,
	})
	if err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}
	return org, nil
}

type searchOrganizationResp struct {
	Id          uint   `json:"id"`
	Name        string `json:"name"`
	Guid        string `json:"guid"`
	Description string `json:"description"`
}

func (m *searchOrganizationResp) TableName() string {
	return models.Organization{}.TableName()
}

func SearchOrganization(c *ctx.ServiceCtx, form *forms.SearchOrganizationForm) (interface{}, e.Error) {
	query := services.QueryUser(c.DB())
	if form.Status != "" {
		query = query.Where("status = ?", form.Status)
	}
	if form.Q != "" {
		qs := "%" + form.Q + "%"
		query = query.Where("name LIKE ?", qs)
	}

	query = query.Order("created_at DESC")
	users := make([]*searchOrganizationResp, 0)
	_ = query.Find(&users)
	return users, nil
}

func UpdateOrganization(c *ctx.ServiceCtx, form *forms.UpdateOrganizationForm) (org *models.Organization, err e.Error) {
	c.AddLogField("action", fmt.Sprintf("update org %d", form.Id))
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

	org, err = services.UpdateOrganization(c.DB(), form.Id, attrs)
	return
}

func DisableOrganization(c *ctx.ServiceCtx, form *forms.DisableOrganizationForm) (interface{}, e.Error) {
	org, er := services.GetOrganizationById(c.DB(), form.Id)
	if er != nil {
		return nil, er
	}

	if org.Status == form.Status {
		return org, nil
	} else if form.Status != models.OrgEnable && form.Status != models.OrgDisable {
		return nil, e.New(e.OrganizationInvalidStatus)
	}

	org, err := services.UpdateOrganization(c.DB(), form.Id, models.Attrs{"status": form.Status})
	if err != nil {
		return nil, err
	}

	return org, nil
}

type organizationDetailResp struct {
	models.Organization
	CreateName string
}

func OrganizationDetail(c *ctx.ServiceCtx, form *forms.DetailOrganizationForm) (resp interface{}, er e.Error) {
	org, err := services.GetOrganizationById(c.DB(), form.Id)
	if err != nil {
		return nil, e.New(e.DBError, http.StatusInternalServerError, err)
	}

	return org, nil
}
