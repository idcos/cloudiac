package models

import "cloudiac/portal/libs/db"

type Policy struct {
	SoftDeleteModel

	GroupId   Id `json:"groupId" gorm:"size:32;comment:策略组ID" example:"lg-c3lcrjxczjdywmk0go90"`
	CreatorId Id `json:"creatorId" gorm:"size:32;not null;创建人" example:"u-c3lcrjxczjdywmk0go90"`

	Name          string `json:"description" gorm:"type:text;comment:名称" example:"ECS分配公网IP"`
	Entry         string `json:"entry" gorm:"comment:rego入口" example:"instanceNoVpc"`
	ReferenceId   string `json:"referenceId" gorm:"not null;size:128;comment:策略ID" example:"iac_aliyun_public_26"`
	Revision      int    `json:"revision" gorm:"default:1;comment:版本" example:"1"`
	Enabled       bool   `json:"enabled" gorm:"default:true;comment:是否全局启用" example:"true"`
	FixSuggestion string `json:"fixSuggestion" gorm:"type:text;comment:策略修复建议" example:"1. 设置 internet_max_bandwidth_out = 0\n 2. 取消设置 allocate_public_ip"`
	Severity      string `json:"severity" gorm:"type:enum('high','medium','low','none');default:'medium';default:medium;comment:严重性" example:"medium"`

	PolicyType   string `json:"policyType" gorm:"comment:云商类型" example:"alicloud"`
	ResourceType string `json:"resourceType" gorm:"comment:资源类型" example:"alicloud_instance"`
	Tags         string `json:"tags" gorm:"comment:标签" example:"security,aliyun"`

	Rego string `json:"rego" gorm:"type:text;comment:rego脚本" example:"package idcos ..."`
}

func (Policy) TableName() string {
	return "iac_policy"
}

func (p *Policy) CustomBeforeCreate(*db.Session) error {
	if p.Id == "" {
		p.Id = NewId("po")
	}
	return nil
}

type PolicyShield struct {
	// todo 是不是直接删除即可
	SoftDeleteModel

	CreatorId Id `json:"creatorId" gorm:"size:32;not null;创建人" example:"u-c3lcrjxczjdywmk0go90"`
	OrgId     Id `json:"orgId" gorm:"not null;size:32;comment:组织ID" example:"org-c3lcrjxczjdywmk0go90"`
	ProjectId Id `json:"projectId" gorm:"default:'';size:32;comment:项目ID" example:"p-c3lcrjxczjdywmk0go90"`
	TplId     Id `json:"tplId" gorm:"size:32;comment:云模板ID" example:"t-c3lcrjxczjdywmk0go90"`
	EnvId     Id `json:"envId" gorm:"size:32;comment:环境ID" example:"e-c3lcrjxczjdywmk0go90"`
	PolicyId  Id `json:"policyId" form:"policyId" gorm:"size:32;comment:策略ID" example:"e-c3lcrjxczjdywmk0go90"`

	Reason string `json:"reason" form:"reason" gorm:"comment:屏蔽说明"`
	Type   string `json:"type" form:"type" gorm:"type:enum('strategy','source');comment:屏蔽类型"`
}

func (PolicyShield) TableName() string {
	return "iac_policy_shield"
}
