package forms

type Params struct {
	Key       string   `json:"key"`
	Value     string   `json:"value"`
	IsSecret  *bool    `json:"isSecret"`
}

type CreateResourceAccountForm struct {
	BaseForm
	Name             string   `form:"name" json:"name" binding:"required,gte=2,lte=32"`
	Description      string   `form:"description" json:"description"`
	Params           []Params `form:"params" json:"params"`
	CtServiceIds     []string `form:"ctServiceIds" json:"ctServiceIds"`
}

type UpdateResourceAccountForm struct {
	BaseForm
	Id              uint     `form:"id" json:"id" binding:"required"`
	Name            string   `form:"name" json:"name" binding:""`
	Description     string   `form:"description" json:"description"`
	Params          string   `form:"params" json:"params"`
	Status          string   `form:"status" json:"status"`
	CtServiceIds    []string `form:"ctServiceIds" json:"ctServiceIds"`
}

type SearchResourceAccountForm struct {
	BaseForm

	Q          string `form:"q" json:"q" binding:""`
	Status     string `form:"status" json:"status"`
}

type DeleteResourceAccountForm struct {
	BaseForm
	Id uint `form:"id" json:"id" binding:"required"`
}