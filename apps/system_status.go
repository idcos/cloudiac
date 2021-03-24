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
	resp := make([]SystemStatusResp, 0)
	serviceInfo,serviceStatus, err :=services.SystemStatusSearch()
	if err != nil {
		return nil ,err
	}

	//构建返回值
	for service, info := range serviceInfo {
		ssr := SystemStatusResp{}
		ssr.Service = service
		for _, serviceInfo := range info {
			for _, status := range serviceStatus {
				if serviceInfo.ID == status.ServiceID {
					ssr.Children = append(ssr.Children, struct {
						ID      string
						Tags    []string `json:"tags" form:"tags" `
						Port    int      `json:"port" form:"port" `
						Address string   `json:"address" form:"address" `
						Status  string   `json:"status" form:"status" `
						Node    string   `json:"node" form:"node" `
						Notes   string   `json:"notes" form:"notes" `
						Output  string   `json:"output" form:"output" `
					}{
						ID:      serviceInfo.ID,
						Tags:    serviceInfo.Tags,
						Port:    serviceInfo.Port,
						Address: serviceInfo.Address,
						Status:  status.Status,
						Node:    status.Node,
						Notes:   status.Notes,
						Output:  status.Output,
					})
				}
			}
		}
		resp = append(resp, ssr)
	}
	return resp,nil
}
