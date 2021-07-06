package forms

import "cloudiac/portal/models"

type CreateVariableForm struct {
	BaseForm

	Scope       string `json:"scope" gorm:"not null;type:enum('org','project','template','env')"`
	Type        string `json:"type" gorm:"not null;type:enum('environment','terraform','ansible')"`
	Name        string `json:"name" gorm:"size:64;not null"`
	Value       string `json:"value" gorm:"default:''"`
	Sensitive   bool   `json:"sensitive" gorm:"default:'0'"`
	Description string `json:"description" gorm:"type:text"`
}

type SearchVariableForm struct {
	PageForm

	Id models.Id `json:"id" form:"id" `
}
