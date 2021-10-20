package forms

import "cloudiac/portal/models"

type SearchVariableGroupForm struct {
	PageForm
	Q string `form:"q" json:"q" binding:""`
}

type CreateVariableGroupForm struct {
	BaseForm
	Name              string                    `json:"name" form:"name"`
	Type              string                    `json:"type" form:"type"`
	VarGroupVariables []VarGroupVariablesCreate `json:"varGroupVariables" form:"varGroupVariables" `
}

type VarGroupVariablesCreate struct {
	Id          string `json:"id" form:"id" `
	Name        string `json:"name" form:"name" `
	Value       string `json:"value" form:"value" `
	Sensitive   bool   `json:"sensitive" form:"sensitive" `
	Description string `json:"description" form:"description" `
}

type UpdateVariableGroupForm struct {
	BaseForm

	Id                models.Id                 `uri:"id"`
	Name              string                    `json:"name" form:"name"`
	Type              string                    `json:"type" form:"type"`
	VarGroupVariables []VarGroupVariablesUpdate `json:"varGroupVariables" form:"varGroupVariables" `
}

type VarGroupVariablesUpdate struct {
	Id          string `json:"id" form:"id" `
	Name        string `json:"name" form:"name" `
	Value       string `json:"value" form:"value" `
	Sensitive   bool   `json:"sensitive" form:"sensitive" `
	Description string `json:"description" form:"description" `
}

type DeleteVariableGroupForm struct {
	BaseForm
	Id models.Id `uri:"id"`
}

type DetailVariableGroupForm struct {
	BaseForm
	Id models.Id `uri:"id"`
}
