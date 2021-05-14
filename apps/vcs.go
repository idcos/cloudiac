package apps

import (
	"cloudiac/consts/e"
	"cloudiac/libs/ctx"
	"cloudiac/models"
	"cloudiac/models/forms"
	"cloudiac/services"
)

func CreateVcs(c *ctx.ServiceCtx, form *forms.CreateVcsForm) (interface{}, e.Error) {
	vcs, err := services.CreateVcs(c.DB(), models.Vcs{
		OrgId: 	  c.OrgId,
		Name:    form.Name,
		VcsType: form.VcsType,
		Status:  form.Status,
		Address: form.Address,
		VcsToken:   form.VcsToken,
	})
	if err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}
	return vcs, nil

}

func UpdateVcs(c *ctx.ServiceCtx, form *forms.UpdateVcsForm) (vcs *models.Vcs, err e.Error) {
	attrs := models.Attrs{}
	if form.HasKey("status") {
		attrs["status"] = form.Status
	}
	if form.HasKey("name") {
		attrs["name"] = form.Name
	}
	if form.HasKey("vcsType") {
		attrs["vcsType"] = form.VcsType
	}
	if form.HasKey("address") {
		attrs["address"] = form.Address
	}
	if form.HasKey("vcsToken") {
		attrs["vcsToken"] = form.VcsToken
	}
	vcs, err = services.UpdateVcs(c.DB(), form.Id, attrs)
	return
}

func SearchVcs(c *ctx.ServiceCtx, form *forms.SearchVcsForm) (interface{}, e.Error) {
	query := services.QueryVcs(c.OrgId, c.DB())
	rs, _ := getPage(query, form, models.Vcs{})
	return rs, nil

}

func DeleteVcs(c *ctx.ServiceCtx, form *forms.DeleteVcsForm) (result interface{}, re e.Error) {
	if err := services.DeleteVcs(c.DB(), form.Id); err != nil {
		return nil, err
	}
	return
}

func ListEnableVcs(c *ctx.ServiceCtx) (interface{}, e.Error) {
	return services.QueryEnableVcs(c.OrgId, c.DB())

}