package models

import "cloudiac/portal/libs/db"

type Policy struct {
	TimedModel

	GroupId   Id `json:"groupId" gorm:"size:32;comment:策略组ID" example:"lg-c3lcrjxczjdywmk0go90"`
	CreatorId Id `json:"creatorId" gorm:"size:32;not null;创建人" example:"u-c3lcrjxczjdywmk0go90"`

	Name          string `json:"description" gorm:"type:text;comment:名称" example:"ECS分配公网IP"`
	ReferenceId   string `json:"referenceId" gorm:"not null;size:128;comment:策略ID" example:"iac_aliyun_public_26"`
	Revision      int    `json:"revision" gorm:"default:1;comment:版本" example:"1"`
	Enabled       bool   `json:"enabled" gorm:"default:true;comment:是否全局启用" example:"true"`
	FixSuggestion string `json:"fixSuggestion" gorm:"type:text;comment:策略修复建议" example:"1. 设置 internet_max_bandwidth_out = 0\n 2. 取消设置 allocate_public_ip"`
	Severity      string `json:"severity" gorm:"type:enum('high','medium','low','none');default:'medium';default:medium;comment:严重性" example:"medium"`

	PolicyType   string `json:"policyType" gorm:"comment:云商类型" example:"alicloud"`
	ResourceType string `json:"resourceType" gorm:"comment:资源类型" example:"alicloud_instance"`
	Category     string `json:"category" gorm:"default:cloudiac;comment:分类" example:"cloudiac"`

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
