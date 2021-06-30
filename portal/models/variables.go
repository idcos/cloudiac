package models

import "cloudiac/portal/libs/db"

type Variable struct {
	BaseModel

	Guid      string `json:"guid" gorm:"size:32;not null;unique"`
	OrgId     uint   `json:"orgId" gorm:"not null"`
	ProjectId uint   `json:"projectId" gorm:"default:'0'"`
	TplId     uint   `json:"tplId" gorm:"default:'0'"`
	EnvId     uint   `json:"envId" gorm:"default:'0'"`
	Scope     string `json:"scope" gorm:"not null;type:enum('org', 'project', 'template', 'env')"`
	Type      string `json:"type" gorm:"not null;type:enum('environment','terraform','ansible')"`

	Name        string `json:"name" gorm:"not null"`
	Value       string `json:"value" gorm:"default:''"`
	Sensitive   bool   `json:"sensitive" gorm:"default:'0'"`
	Description string `json:"description" gorm:"type:text"`
}

func (Variable) TableName() string {
	return "iac_variable"
}

func (v Variable) Migrate(sess *db.Session) error {
	// 变量名在各 scope 下唯一
	// 注意这些 id 字段需要默认设置为 0，否则唯一索引可能会因为存在 null 值而不生效
	return v.AddUniqueIndex(sess, "unique__variable__name",
		"org_id", "project_id", "tpl_id", "env_id", "name")
}
