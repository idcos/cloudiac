package billcollect

import (
	"cloudiac/portal/consts"
	"cloudiac/portal/models"
	"github.com/pkg/errors"
)

type ResourceCost struct {
	ProductCode    string `json:"productCode"`    // 产品类型
	InstanceId     string `json:"instanceId"`     // 实例id
	InstanceConfig string `json:"instanceConfig"` // 实例配置
	PretaxAmount   float32  `json:"pretaxAmount"`   // 应付金额
	Region         string `json:"region"`         // 区域
	Currency       string `json:"currency"`       // 币种
	Cycle          string `json:"cycle"`          // 账单月
	Provider       string `json:"provider"`
}


type ClintIface interface {
	// GetResourceMonthCost 获取月账单数据
	// param billingCycle 账单采集周期
	GetResourceMonthCost(billingCycle string) ([]ResourceCost, error)

	// GetResourceDayCost 获取日账单数据
	// param billingCycle 账单采集周期
	GetResourceDayCost(billingCycle string) ([]ResourceCost, error)
}

func GetBillInstance(vg *models.VariableGroup)(ClintIface,error){
	switch vg.Provider {
	case consts.BillCollectAli :
		return newAliBillInstance(vg)
	default:
		return nil,errors.New("bill type not exist")

	}
}
