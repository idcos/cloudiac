// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package models

import (
	"cloudiac/portal/libs/db"
	"database/sql/driver"
	"encoding/json"
)

type VariableBody struct {
	Scope       string `yaml:"scope" json:"scope" gorm:"not null;type:enum('org','template','project','env')"`
	Type        string `yaml:"type" json:"type" gorm:"not null;type:enum('environment','terraform','ansible')"`
	Name        string `yaml:"name" json:"name" gorm:"size:64;not null"`
	Value       string `yaml:"value" json:"value" gorm:"type:text"`
	Sensitive   bool   `yaml:"sensitive" json:"sensitive,omitempty" gorm:"default:false"`
	Description string `yaml:"description" json:"description,omitempty" gorm:"type:text"`

	// 继承关系依赖数据创建枚举的顺序，后续新增枚举值时请按照新的继承顺序增加
	Options StrSlice `yaml:"options" json:"options" gorm:"type:json"` // 可选值列表
}

func (v *VariableBody) Key() string {
	return v.Type + ":" + v.Name
}

type Variable struct {
	BaseModel
	VariableBody

	OrgId     Id `json:"orgId" gorm:"size:32;not null"`
	ProjectId Id `json:"projectId" gorm:"size:32;default:''"`
	TplId     Id `json:"tplId" gorm:"size:32;default:''"`
	EnvId     Id `json:"envId" gorm:"size:32;default:''"`
}

func (Variable) NewId() Id {
	return NewId("var")
}

func (Variable) TableName() string {
	return "iac_variable"
}

func (v Variable) MarshalJSON() ([]byte, error) {
	type Tv Variable
	tv := Tv(v)
	if tv.Sensitive {
		tv.Value = ""
	}
	return json.Marshal(tv)
}

func (v Variable) Migrate(sess *db.Session) error {
	// 变量名在各 scope 下唯一
	// 注意这些 id 字段需要默认设置为 ''，否则联合唯一索引可能会因为存在 null 值而不生效
	if err := v.AddUniqueIndex(sess, "unique__variable__name",
		"org_id", "project_id", "tpl_id", "env_id", "name(32)", "type"); err != nil {
		return err
	}
	if err := sess.ModifyModelColumn(&v, "value"); err != nil {
		return err
	}
	return nil
}

type VariableGroup struct {
	TimedModel
	Name      string            `json:"name" gorm:"size:64;not null"`
	Type      string            `json:"type" gorm:"not null;type:enum('environment','terraform')"`
	CreatorId Id                `json:"creatorId" gorm:"size:32;not null;comment:创建人" example:"u-c3ek0co6n88ldvq1n6ag"`
	OrgId     Id                `json:"orgId" gorm:"size:32;not null"`
	Variables VarGroupVariables `json:"variables" gorm:"type:json;null;comment:变量组下的变量"`

	CostCounted bool   `json:"costCounted" gorm:"default:false;comment:是否开启费用统计"` // 是否开启费用统计
	Provider    string `json:"provider" gorm:"comment:资源供应平台名称"`                  // 资源供应平台名称
}

func (VariableGroup) TableName() string {
	return "iac_variable_group"
}

func (VariableGroup) NewId() Id {
	return NewId("vg")
}

func (vg VariableGroup) MarshalJSON() ([]byte, error) {
	type Tvg VariableGroup
	tvg := Tvg(vg)
	for i, v := range tvg.Variables {
		if v.Sensitive {
			tvg.Variables[i].Value = ""
		}
	}
	return json.Marshal(tvg)
}

func (v VariableGroup) Migrate(sess *db.Session) error {
	if err := v.AddUniqueIndex(sess, "unique__org__variable_group_name",
		"org_id", "name(32)"); err != nil {
		return err
	}
	return nil
}

type VarGroupVariables []VarGroupVariable

func (v VarGroupVariables) Value() (driver.Value, error) {
	return MarshalValue(v)
}

func (v *VarGroupVariables) Scan(value interface{}) error {
	return UnmarshalValue(value, v)
}

type VarGroupVariable struct {
	Id          string `json:"id" form:"id" `
	Name        string `json:"name" form:"name" `
	Value       string `json:"value" form:"value" `
	Sensitive   bool   `json:"sensitive" form:"sensitive" `
	Description string `json:"description" form:"description" `
}

//VariableGroupRel 变量组与实例的关联表
type VariableGroupRel struct {
	AbstractModel
	VarGroupId Id     `json:"varGroupId" gorm:"size:32;not null"`
	ObjectType string `json:"objectType" gorm:"not null; type:enum('org','template','project','env')"`
	ObjectId   Id     `json:"objectId" gorm:"size:32;not null"`
}

func (VariableGroupRel) TableName() string {
	return "iac_variable_group_rel"
}

//VariableGroupProjectRel 变量组与项目的关联表
/*
如果用户没有为变量指定绑定的项目，则默认表示绑定到所有项目（包括之后新创建的项目）。
在该表中使用 ProjectId  为 "" 的记录表示绑定到所有项目。
*/
type VariableGroupProjectRel struct {
	AbstractModel
	VarGroupId Id `json:"varGroupId" gorm:"uniqueIndex:idx_var_group_project;size:32;not null"`
	ProjectId  Id `json:"projectId" gorm:"uniqueIndex:idx_var_group_project;size:32;not null"`
}

func (VariableGroupProjectRel) TableName() string {
	return "iac_variable_group_project_rel"
}
