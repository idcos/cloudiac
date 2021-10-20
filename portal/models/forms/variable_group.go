package forms

import "cloudiac/portal/models"

type SearchVariableGroupForm struct {
	PageForm
	Q string `form:"q" json:"q" binding:""`
}

type CreateVariableGroupForm struct {
	BaseForm
	Name      string                    `json:"name" form:"name"`
	Type      string                    `json:"type" form:"type"`
	Variables []models.VarGroupVariable `json:"variables" form:"variables" `
}

type UpdateVariableGroupForm struct {
	BaseForm

	Id        models.Id                 `uri:"id"`
	Name      string                    `json:"name" form:"name"`
	Type      string                    `json:"type" form:"type"`
	Variables []models.VarGroupVariable `json:"variables" form:"variables" `
}

type DeleteVariableGroupForm struct {
	BaseForm
	Id models.Id `uri:"id"`
}

type DetailVariableGroupForm struct {
	BaseForm
	Id models.Id `uri:"id"`
}
