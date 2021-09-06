package models

import "cloudiac/portal/libs/db"

// 策略屏蔽机制
// 1. 策略可以禁用 -> policy
// 2. 策略组可以禁用 -> policy_group
// 3. 云模板可以禁用 -> policy_suppress
// 4. 环境可以禁用 -> policy_suppress

// 云模板扫描的时候检查 policy suppress 是否被屏蔽

type PolicySuppress struct {
	TimedModel

	CreatorId Id `json:"creatorId" gorm:"size:32;not null;创建人" example:"u-c3lcrjxczjdywmk0go90"`
	OrgId     Id `json:"orgId" gorm:"not null;size:32;comment:组织ID" example:"org-c3lcrjxczjdywmk0go90"`
	ProjectId Id `json:"projectId" gorm:"default:'';size:32;comment:项目ID" example:"p-c3lcrjxczjdywmk0go90"`
	TplId     Id `json:"tplId" gorm:"size:32;comment:云模板ID" example:"t-c3lcrjxczjdywmk0go90"`
	EnvId     Id `json:"envId" gorm:"size:32;comment:环境ID" example:"e-c3lcrjxczjdywmk0go90"`
	PolicyId  Id `json:"policyId" form:"policyId" gorm:"size:32;not null;comment:策略ID" example:"e-c3lcrjxczjdywmk0go90"`

	Reason string `json:"reason" form:"reason" gorm:"comment:屏蔽说明"`
	Type   string `json:"type" form:"type" gorm:"type:enum('policy','source');comment:屏蔽类型"`
}

func (PolicySuppress) TableName() string {
	return "iac_policy_suppress"
}

func (p *PolicySuppress) CustomBeforeCreate(*db.Session) error {
	if p.Id == "" {
		p.Id = NewId("pos")
	}
	return nil
}
