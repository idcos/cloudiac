// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package apps

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
	"fmt"
	"net/http"
)

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

func SystemStatusSearch() (interface{}, e.Error) {
	resp := make([]*SystemStatusResp, 0)
	serviceResp := make(map[string]*SystemStatusResp)
	IdInfo, serviceStatus, serviceList, err := services.SystemStatusSearch()
	if err != nil {
		return nil, err
	}

	//构建返回值
	for _, service := range serviceList {
		serviceResp[service] = &SystemStatusResp{
			Service: service,
		}
	}

	for _, id := range IdInfo {
		serviceResp[id.Service].Children = append(serviceResp[id.Service].Children, struct {
			ID      string
			Tags    []string `json:"tags" form:"tags" `
			Port    int      `json:"port" form:"port" `
			Address string   `json:"address" form:"address" `
			Status  string   `json:"status" form:"status" `
			Node    string   `json:"node" form:"node" `
			Notes   string   `json:"notes" form:"notes" `
			Output  string   `json:"output" form:"output" `
		}{
			ID:      id.ID,
			Tags:    id.Tags,
			Port:    id.Port,
			Address: id.Address,
			Status:  serviceStatus[id.ID].Status,
			Node:    serviceStatus[id.ID].Node,
			Notes:   serviceStatus[id.ID].Notes,
			Output:  serviceStatus[id.ID].Output,
		})
	}

	for _, service := range serviceResp {
		resp = append(resp, service)
	}

	return resp, nil
}

func ConsulKVSearch(key string) (interface{}, e.Error) {
	return services.ConsulKVSearch(key)
}

func RunnerSearch() (interface{}, e.Error) {
	return services.RunnerSearch()
}

func ConsulTagUpdate(c *ctx.ServiceContext, form forms.ConsulTagUpdateForm) (interface{}, e.Error) {
	// 检查是否有修改tags的权限
	if !c.IsSuperAdmin {
		return nil, e.New(e.PermissionDeny, fmt.Errorf("super admin required"), http.StatusForbidden)
	}

	//将修改后的tag存到consul中
	if err := services.ConsulKVSave(form.ServiceId, form.Tags); err != nil {
		return nil, err
	}
	//根据serviceId查询在consul中保存的数据
	agentService, err := services.ConsulServiceInfo(form.ServiceId)
	if err != nil {
		return nil, err
	}
	//重新注册
	if err := services.ConsulServiceRegistered(agentService, form.Tags); err != nil {
		return nil, err
	}
	return nil, nil
}

func RunnerTags() (interface{}, e.Error) {
	tags, err := services.SystemRunnerTags()
	if err != nil {
		return nil, err
	}

	return &RunnerTagsResp{
		Tags: tags,
	}, nil
}
