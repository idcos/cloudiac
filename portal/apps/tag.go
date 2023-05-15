package apps

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
)

func UpdateTag(c *ctx.ServiceContext, form *forms.UpdateTagForm) (interface{}, e.Error) {
	attr := models.Attrs{}
	attr["value"] = form.Value
	return services.UpdateTag(c.DB(), form.TagId, form.Id, attr)
}

func DeleteTag(c *ctx.ServiceContext, form *forms.DeleteTagForm) (interface{}, e.Error) {
	return services.DeleteTag(c.DB(), form.TagId, form.Id)
}

func CreateTag(c *ctx.ServiceContext, form *forms.CreateTagForm) (interface{}, e.Error) {
	tx := c.Tx()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	services.CreateTag()
	return nil, nil
}

func SearchTag(c *ctx.ServiceContext, form * forms.CreateTagForm) (interface{}, e.Error) {}

