package apps

import (
	"cloudiac/consts/e"
	"cloudiac/services"
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

func SystemStatusSearch() (interface{}, e.Error) {
	resp := make([]*SystemStatusResp, 0)
	serviceResp := make(map[string]*SystemStatusResp, 0)
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
