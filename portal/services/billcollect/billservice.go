// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package billcollect

import (
	"cloudiac/portal/consts"
	"cloudiac/portal/models"
	"cloudiac/utils"
	"cloudiac/utils/logs"
	"fmt"
)

type ResourceCost struct {
	ProductCode    string  `json:"productCode"`    // 产品类型
	InstanceId     string  `json:"instanceId"`     // 实例id
	InstanceConfig string  `json:"instanceConfig"` // 实例配置
	PretaxAmount   float32 `json:"pretaxAmount"`   // 应付金额
	Region         string  `json:"region"`         // 区域
	Currency       string  `json:"currency"`       // 币种
	Cycle          string  `json:"cycle"`          // 账单月
	Provider       string  `json:"provider"`
}

type BillProvider interface {
	// Provider 返回云商名称, 如 aws, alicloud 等
	Provider() string

	// ParseMonthBill 解析账单数据
	// param billingCycle 账单采集周期
	ParseMonthBill(billingCycle string) (map[string]ResourceCost, []string, []models.BillData, error)

	//// GetResourceMonthCost 获取月账单数据
	//// param billingCycle 账单采集周期
	//GetResourceMonthCost(billingCycle string) ([]*bssopenapi20171214.QueryInstanceBillResponseBodyDataItemsItem, error)

	// GetResourceDayCost 获取日账单数据
	// param billingCycle 账单采集周期
	GetResourceDayCost(billingCycle string) ([]ResourceCost, error)
}

func GetBillProvider(vg *models.VariableGroup) (BillProvider, error) {
	switch vg.Provider {
	case consts.BillCollectAli:
		return NewAlicloudBillProvider(vg)
	default:
		logs.Get().Errorf("unsupported provider %s", vg.Provider)
		return nil, fmt.Errorf("unsupported provider %s", vg.Provider)

	}
}

func parseResourceAccount(provider string, vars models.VarGroupVariables) map[string]string {
	resp := make(map[string]string)
	for _, v := range vars {
		if utils.ArrayIsExistsStr(consts.BillProviderResAccount[provider], v.Name) {
			if v.Sensitive {
				value, err := utils.DecryptSecretVarForce(v.Value)
				if err != nil {
					logs.Get().Errorf("get resource as sk error: %s", err)
					return nil
				}
				resp[v.Name] = value
				continue
			}
			resp[v.Name] = v.Value
		}
	}
	return resp
}
