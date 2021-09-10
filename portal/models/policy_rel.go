package models

import "cloudiac/portal/libs/db"

const (
	PolicyRelScopeEnv = "env"
	PolicyRelScopeTpl = "template"
)

// 策略关系表存储四种记录
// 1. 策略组与环境  tplId, groupId
// 2. 策略与环境    envId, groupId
// 3. 云模板启用    tplId, scope = '', enabled = true
// 4. 环境启用      envId, groupId = '', enabled = true

type PolicyRel struct {
	AutoUintIdModel

	OrgId     Id `json:"orgId" gorm:"not null;size:32;comment:组织" example:"org-c3lcrjxczjdywmk0go90"`
	ProjectId Id `json:"projectId" gorm:"default:'';size:32;comment:项目ID" example:"p-c3lcrjxczjdywmk0go90"`

	GroupId Id     `json:"groupId" gorm:"size:32;comment:策略组ID" example:"lg-c3lcrjxczjdywmk0go90"`
	TplId   Id     `json:"tplId" gorm:"default:'';size:32;comment:云模板ID" example:"tpl-c3lcrjxczjdywmk0go90"`
	EnvId   Id     `json:"envId" gorm:"default:'';size:32;comment:环境ID" example:"env-c3lcrjxczjdywmk0go90"`
	Scope   string `json:"scope" gorm:"not null;enums:('template','env', 'group');comment:绑定范围" enums:"template,env" example:"env"`

	Enabled bool `json:"enabled" gorm:"default:false;comment:启用状态" example:"true"` // 是否启用环境/策略扫描，无关联记录也表示不启用扫描
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
