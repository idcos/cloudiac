package models

import "cloudiac/portal/libs/db"

type VariableBody struct {
	// 继承关系依赖数据创建枚举的顺序，后续新增枚举值时请按照新的继承顺序增加
	Scope       string `json:"scope" gorm:"not null;type:enum('org','template','project','env')"`
	Type        string `json:"type" gorm:"not null;type:enum('environment','terraform','ansible')"`
	Name        string `json:"name" gorm:"size:64;not null"`
	Value       string `json:"value" gorm:"default:''"`
	Sensitive   bool   `json:"sensitive,omitempty" gorm:"default:'0'"`
	Description string `json:"description,omitempty" gorm:"type:text"`
}

type Variable struct {
	TimedModel
	VariableBody

	OrgId     Id `json:"orgId" gorm:"size:32;not null"`
	ProjectId Id `json:"projectId" gorm:"size:32;default:'0'"`
	TplId     Id `json:"tplId" gorm:"size:32;default:'0'"`
	EnvId     Id `json:"envId" gorm:"size:32;default:'0'"`
}

func (Variable) TableName() string {
	return "iac_variable"
}

func (v Variable) Migrate(sess *db.Session) error {
	// 变量名在各 scope 下唯一
	// 注意这些 id 字段需要默认设置为 0，否则联合唯一索引可能会因为存在 null 值而不生效
	return v.AddUniqueIndex(sess, "unique__variable__name",
		"org_id", "project_id", "tpl_id", "env_id", "name(32)")
}
