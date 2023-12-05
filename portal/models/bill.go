// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package models

type Bill struct {
	AutoUintIdModel

	OrgId          Id      `json:"orgId" gorm:"size:32;not null"`            // 组织ID
	ProjectId      Id      `json:"projectId" gorm:"size:32;not null"`        // 项目ID
	EnvId          Id      `json:"envId" gorm:"size:32;not null"`            // 模板ID
	VgId           Id      `json:"vgId"  gorm:"size:32;not null"`            // 资源账号id
	ProductCode    string  `json:"productCode" gorm:"not null"`              // 产品类型
	InstanceId     string  `json:"instanceId" gorm:"not null"`               // 实例id
	InstanceConfig string  `json:"instanceConfig" gorm:"type:text;not null"` // 实例配置
	PretaxAmount   float32 `json:"pretaxAmount" gorm:"not null"`             // 应付金额
	Region         string  `json:"region" gorm:"not null"`                   // 区域
	Currency       string  `json:"currency" gorm:"not null"`                 // 币种
	Cycle          string  `json:"cycle" gorm:"not null"`                    // 账单月
	Provider       string  `json:"provider" gorm:"not null"`
}

func (Bill) TableName() string {
	return "iac_bill"
}

type BillData struct {
	AutoUintIdModel

	Provider   string   `json:"provider"`
	InstanceId string   `json:"instanceId"`
	Attrs      ResAttrs `json:"attrs,omitempty" gorm:"type:text"`
}

func (BillData) TableName() string {
	return "iac_bill_data"
}
