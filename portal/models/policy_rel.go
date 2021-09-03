package models

import "cloudiac/portal/libs/db"

const (
	PolicyRelScopeEnv = "environment"
	PolicyRelScopeTpl = "template"
)

type PolicyRel struct {
	AutoUintIdModel

	OrgId     Id `json:"orgId" gorm:"not null;size:32;comment:组织" example:"org-c3lcrjxczjdywmk0go90"`
	ProjectId Id `json:"projectId" gorm:"default:'';size:32;comment:项目ID" example:"p-c3lcrjxczjdywmk0go90"`

	GroupId Id     `json:"groupId" gorm:"not null;size:32;comment:策略组ID" example:"lg-c3lcrjxczjdywmk0go90"`
	TplId   Id     `json:"tplId" gorm:"default:'';size:32;comment:云模板ID" example:"tpl-c3lcrjxczjdywmk0go90"`
	EnvId   Id     `json:"envId" gorm:"default:'';size:32;comment:环境ID" example:"env-c3lcrjxczjdywmk0go90"`
	Scope   string `json:"scope" gorm:"not null;enums:('template','env');comment:绑定范围" example:"env"`
}

func (PolicyRel) TableName() string {
	return "iac_policy_rel"
}

func (r PolicyRel) Migrate(sess *db.Session) error {
	if err := r.AddUniqueIndex(sess, "unique__group__tpl__env",
		"group_id", "tpl_id", "env_id"); err != nil {
		return err
	}
	return nil
}
