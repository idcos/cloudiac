package models

import "cloudiac/portal/libs/db"

type ProjectTemplate struct {
	AutoUintIdModel
	ProjectId  Id `json:"projectId" gorm:"not null"`
	TemplateId Id `json:"templateId" gorm:"not null"`
}

func (ProjectTemplate) TableName() string {
	return "iac_project_template"
}

func (u ProjectTemplate) Migrate(sess *db.Session) error {
	return u.AddUniqueIndex(sess, "unique__project__template", "project_id", "template_id")
}
