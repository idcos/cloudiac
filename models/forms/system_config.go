package forms

type SearchSystemConfigForm struct {
	BaseForm

	Q string `form:"q" json:"q" binding:""`
}

type UpdateSystemConfigForm struct {
	BaseForm
	Id          uint   `form:"id" json:"id" binding:"required"`
	Name        string `form:"name" json:"name" binding:"required"`
	Value       string `form:"value" json:"value" binding:"required"`
	Description string `form:"description" json:"description"`
}
