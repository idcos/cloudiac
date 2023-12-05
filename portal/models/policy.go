// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package models

import (
	"cloudiac/common"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"strings"
	"unicode/utf8"
)

const MaxTagSize = 16

type Policy struct {
	SoftDeleteModel

	OrgId     Id `json:"orgId" gorm:"size:32;comment:组织ID" example:"org-c3lcrjxczjdywmk0go90"`
	GroupId   Id `json:"groupId" gorm:"size:32;comment:策略组ID" example:"lg-c3lcrjxczjdywmk0go90"`
	CreatorId Id `json:"creatorId" gorm:"size:32;not null;创建人" example:"u-c3lcrjxczjdywmk0go90"`

	Name          string `json:"name" gorm:"type:text;comment:名称" example:"ECS分配公网IP"`
	RuleName      string `json:"ruleName" gorm:"comment:rego规则名称" example:"instanceNoVpc"`
	ReferenceId   string `json:"referenceId" gorm:"not null;size:128;comment:策略ID" example:"iac_aliyun_public_26"`
	Revision      int    `json:"revision" gorm:"default:1;comment:版本" example:"1"`
	Enabled       bool   `json:"enabled" gorm:"default:true;comment:是否全局启用" example:"true"`
	FixSuggestion string `json:"fixSuggestion" gorm:"type:text;comment:策略修复建议" example:"1. 设置 internet_max_bandwidth_out = 0\n 2. 取消设置 allocate_public_ip"`
	Severity      string `json:"severity" gorm:"default:'medium';default:medium;comment:严重性" example:"medium"` // type:enum('high','medium','low');

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

func (p *Policy) Validate() error {
	return p.ValidateAttrs(Attrs{
		"tags": p.Tags,
	})
}

func (p *Policy) ValidateAttrs(attrs Attrs) error {
	for k, v := range attrs {
		switch db.ToColName(k) {
		case "tags":
			for i, tag := range strings.Split(v.(string), ",") {
				// 限制只允许有 10 个 tag
				if i >= 10 {
					return e.New(e.TagTooMuch)
				}

				if utf8.RuneCountInString(tag) > MaxTagSize {
					return e.New(e.TagTooLong)
				}
			}
		}
	}
	return nil
}

/*
PolicyStatusConversion 转换规则
开启状态	scan_task status	扫描状态	前端状态	policyStatus
未开启	（无）	未开启	未开启	disable
已开启	-（无扫描记录）	开启未检测	未检测	enable
已开启	pending	检测中	“单鱼转”动画	pending
已开启	passed	检测通过	合规	passed
已开启	violated	检测不通过	不合规	violated
已开启	failed	检测失败	不合规	violated
已开启	violated	检测不通过+检测失败	不合规	violated
*/
func PolicyStatusConversion(status string, enable bool) string {
	if !enable {
		return common.PolicyStatusDisable
	}

	switch status {
	case common.PolicyStatusPending:
		return common.PolicyStatusPending
	case common.PolicyStatusPassed:
		return common.PolicyStatusPassed
	case common.PolicyStatusViolated:
		return common.PolicyStatusViolated
	case common.PolicyStatusFailed:
		return common.PolicyStatusViolated
	default:
		return common.PolicyStatusEnable
	}
}
