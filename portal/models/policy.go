package models

type Policy struct {
	TimedModel
	GroupId   Id `json:"group_id" gorm:"size:32;comment:策略组ID" example:"lg-c3lcrjxczjdywmk0go90"`
	CreatorId Id `json:"creator_id" gorm:"size:32;not null;创建人" example:"u-c3lcrjxczjdywmk0go90"`

	Name          string `json:"description" gorm:"type:text;comment:名称" example:"ECS分配公网IP"`
	ReferenceId   string `json:"reference_id" gorm:"not null;size:128;comment:策略ID" example:"iac_aliyun_public_26"`
	Revision      string `json:"revision" gorm:"default:1;comment:版本" example:"1"`
	Enabled       bool   `json:"enabled" gorm:"default:true;comment:是否全局启用" example:"true"`
	FixSuggestion string `json:"fix_suggestion" gorm:"type:text;comment:策略修复建议" example:"1. 设置 internet_max_bandwidth_out = 0\n 2. 取消设置 allocate_public_ip"`
	Severity      string `json:"severity" gorm:"type:enum('high','medium','low','none');default:'medium';default:medium;comment:严重性" example:"medium"`

	PolicyType   string `json:"policy_type" gorm:"comment:云商类型" example:"alicloud"`
	ResourceType string `json:"resource_type" gorm:"comment:资源类型" example:"alicloud_instance"`
	Category     string `json:"category" gorm:"comment:分类"`

	Rego string `json:"rego" gorm:"type:text;comment:rego脚本" example:"package idcos ..."`
}

func (Policy) TableName() string {
	return "iac_policy"
}

func (l *Policy) GetId() Id {
	if l.Id == "" {
		return NewId("po")
	}
	return l.Id
}
