// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package resps

import "cloudiac/portal/models"

type SearchSystemConfigResp struct {
	Id          models.Id `json:"id"`
	Name        string    `json:"name"`
	Value       string    `json:"value"`
	Description string    `json:"description"`
}

func (m *SearchSystemConfigResp) TableName() string {
	return models.SystemCfg{}.TableName()
}

type RegistryAddrResp struct {
	RegistryAddrFromDB  string `json:"registryAddrDB"`
	RegistryAddrFromCfg string `json:"registryAddrCfg"`
}

type SystemStatusResp struct {
	Service  string `json:"service" form:"service" `
	Children []struct {
		ID      string
		Tags    []string `json:"tags" form:"tags" `
		Port    int      `json:"port" form:"port" `
		Address string   `json:"address" form:"address" `
		Status  string   `json:"status" form:"status" `
		Node    string   `json:"node" form:"node" `
		Notes   string   `json:"notes" form:"notes" `
		Output  string   `json:"output" form:"output" `
	}
	//Passing  uint64 `json:"passing" form:"passing" `
	//Critical uint64 `json:"critical" form:"critical" `
	//Warn     uint64 `json:"warn" form:"warn" `
}

type RunnerTagsResp struct {
	Tags []string `json:"tags"`
}

type SystemSwitchesStatusResp struct {
	AbortStatus     bool `json:"abortStatus"` // 与 enableAbortTask 值相同(兼容处理)
	EnableAbortTask bool `json:"enableAbortTask"`

	EnableRegister bool `json:"enableRegister"`
	EnableLdap     bool `json:"enableLdap"`
}

type UserEmailStatus struct {
	Email        string `json:"email"`
	ActiveStatus string `json:"activeStatus"`
}
