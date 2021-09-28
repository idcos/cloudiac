package models

import "cloudiac/portal/libs/db"

type PolicySuppress struct {
	TimedModel

	CreatorId  Id     `json:"creatorId" gorm:"size:32;not null;创建人" example:"u-c3lcrjxczjdywmk0go90"`                                                                        // 操作人
	OrgId      Id     `json:"orgId" gorm:"not null;size:32;comment:组织ID" example:"org-c3lcrjxczjdywmk0go90"`                                                                 // 组织ID
	ProjectId  Id     `json:"projectId" gorm:"default:'';size:32;comment:项目ID" example:"p-c3lcrjxczjdywmk0go90"`                                                             // 项目ID
	TargetId   Id     `json:"targetId" gorm:"uniqueIndex:unique__policy__target;size:32;not null;comment:目标ID" example:"tpl-c3lcrjxczjdywmk0go90"`                           // 屏蔽ID，根据屏蔽类型可以为环境ID、云模板ID或者策略ID
	TargetType string `json:"targetType" gorm:"not null;comment:屏蔽目标类型;type:enum('env','template','policy')" enums:"env,template,policy" example:"env-c3lcrjxczjdywmk0go90"` // 屏蔽目标类型：env环境，template云模板，policy策略
	PolicyId   Id     `json:"policyId" gorm:"uniqueIndex:unique__policy__target;size:32;not null;comment:策略ID" example:"po-c3lcrjxczjdywmk0go90"`                            // 策略ID
	Reason     string `json:"reason" gorm:"comment:屏蔽说明" example:"测试环境不检测此策略"`                                                                                               // 屏蔽原因
	Type       string `json:"type" gorm:"type:enum('policy','source');comment:屏蔽类型" enums:"policy,source" example:"source"`                                                  // 屏蔽类型：policy按策略屏蔽，source按来源屏蔽
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
