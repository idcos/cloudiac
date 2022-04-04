// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package resps

type EnvUnLockConfirmResp struct {
	AutoDestroyPass bool `json:"autoDestroyPass"`
}

type EnvCostTypeStatResp struct {
	ResType string  `json:"resType"`
	Amount  float32 `json:"amount"`
}

type EnvCostTrendStatResp struct {
	Date   string  `json:"date"`
	Amount float32 `json:"amount"`
}

type EnvCostDetail struct {
	ResType      string  `json:"resType"`
	ResAddr      string  `json:"resAddr"`
	InstanceId   string  `json:"instanceId"` // 实例id
	CurMonthCost float32 `json:"curMonthCost"`
	TotalCost    float32 `json:"totalCost"`
}

type EnvStatisticsResp struct {
	CostTypeStat  []EnvCostTypeStatResp  `json:"costTypeStat"`
	CostTrendStat []EnvCostTrendStatResp `json:"costTrendStat"`
	CostList      []EnvCostDetail        `json:"costList"`
}
