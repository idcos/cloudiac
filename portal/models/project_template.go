package models

type ProjectTemplate struct {
	ProjectId	string  `json:"projectId" gorm:"not null"`
	TemplateId  string  `json:"templateId" gorm:"not null"`
}

func (ProjectTemplate) TableName() string {
	return "iac_project_template"
}
