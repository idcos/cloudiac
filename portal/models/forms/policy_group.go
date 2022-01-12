package forms

type SearchRegistryPgForm struct {
	PageForm

	Q string `json:"q" form:"q"`
}

type SearchRegistryPgVersForm struct {
	BaseForm

	Namespace string `json:"ns" form:"ns" binding:"required"`
	GroupName string `json:"gn" form:"gn" binding:"required"`
}
