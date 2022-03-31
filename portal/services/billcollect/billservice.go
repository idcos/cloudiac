package billcollect

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

type BillIface interface {
	Clint() (ClintIface, error)
}

type ClintIface interface {
	GetResourceCost() ([]ResourceCost, error)
}
