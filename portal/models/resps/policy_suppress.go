// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package resps

import "cloudiac/portal/models"

type PolicySuppressResp struct {
	models.PolicySuppress
	TargetName string `json:"targetName"` // 检查目标
	Creator    string `json:"creator"`    // 操作人
}

func (PolicySuppressResp) TableName() string {
	return "s"
}

type PolicySuppressSourceResp struct {
	TargetId   models.Id `json:"targetId" example:"env-c3lcrjxczjdywmk0go90"`   // 屏蔽源ID
	TargetType string    `json:"targetType" enums:"env,template" example:"env"` // 源类型：env环境, template云模板
	TargetName string    `json:"targetName" example:"测试环境"`                     // 名称
}

func (PolicySuppressSourceResp) TableName() string {
	return "iac_policy_suppress"
}
