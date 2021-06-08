package forms

type CreateTemplateLibraryForm struct {
	BaseForm
	Id   uint   `json:"id" form:"id" `
	Name string `json:"name" form:"name" `
}

type SearchTemplateLibraryForm struct {
	BaseForm
}
