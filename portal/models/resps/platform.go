// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package resps

type BaseDataCount struct {
	Total  int64 `json:"total"`
	Active int64 `json:"active"`
}

type PfBasedataResp struct {
	OrgCount     BaseDataCount `json:"orgCount"`
	ProjectCount BaseDataCount `json:"projectCount"`
	EnvCount     BaseDataCount `json:"envCount"`
	StackCount   BaseDataCount `json:"stackCount"`
	UserCount    BaseDataCount `json:"userCount"`
}

type PfProEnvStatResp struct {
	ProName string `json:"proName"`
	Count   int64  `json:"count"`
}
