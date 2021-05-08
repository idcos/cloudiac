package apps

import (
	"cloudiac/consts/e"
	"cloudiac/libs/ctx"
	"cloudiac/models"
	"cloudiac/models/forms"
	"cloudiac/services"
	"fmt"
)

func CreateVcs(c *ctx.ServiceCtx, form *forms.CreateVcsForm) (interface{}, e.Error) {
	vcs, err := services.CreateVcs(c.DB(), models.Vcs{
		Name:    form.Name,
		VcsType: form.VcsType,
		Status:  form.Status,
		Address: form.Address,
		Token:   form.VcsToken,
	})
	if err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}
	return vcs, nil

}

func UpdateVcs(c *ctx.ServiceCtx, form *forms.UpdateVcsForm) (vcs *models.Vcs, err e.Error) {
	attrs := models.Attrs{}
	if form.Id == 0 {
		return nil, e.New(e.BadRequest, fmt.Errorf("missing 'id'"))
	}
	if form.HasKey("status") {
		attrs["status"] = form.Status
	}

	vcs, err = services.UpdateVcs(c.DB(), form.Id, attrs)
	return
}

func SearchVcs(c *ctx.ServiceCtx, form *forms.SearchVcsForm) (interface{}, e.Error) {
	query := services.QueryVcs(c.DB())
	// TODO 这个查的逻辑，是直接返回这个表中所有的数据嘛，没看到limit的代码。
	rs, _ := getPage(query, form, models.Vcs{})
	return rs, nil

}

func DeleteVcs(c *ctx.ServiceCtx, form *forms.DeleteVcsForm) (result interface{}, re e.Error) {
	// TODO 为什么增加 LogField 字段
	// TODO 根据自增ID 去删除
	//c.AddLogField("action", fmt.Sprintf("delete token %d", form.Id))
	if err := services.DeleteVcs(c.DB(), form.Id); err != nil {
		return nil, err
	}
	return
}
