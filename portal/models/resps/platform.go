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
	Provider string `json:"provider"`
	Count    int64  `json:"count"`
}

type PfProResStatResp struct {
	Provider string `json:"provider"`
	Count    int64  `json:"count"`
}

type PfResTypeStatResp struct {
	ResType string `json:"resType"`
	Count   int64  `json:"count"`
}
