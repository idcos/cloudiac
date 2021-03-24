package services

import (
	"cloudiac/configs"
	"cloudiac/consts/e"
	"github.com/hashicorp/consul/api"
)

func SystemStatusSearch() ([]api.AgentService, map[string]api.AgentCheck, []string, e.Error) {
	serviceList := make([]string, 0)
	IdInfo := make([]api.AgentService, 0)
	serviceStatus := make(map[string]api.AgentCheck, 0)
	conf := configs.Get()
	config := api.DefaultConfig()
	config.Address = conf.Consul.Address

	client, err := api.NewClient(config)
	if err != nil {
		return nil, nil, nil, e.New(e.ConsulConnError, err)
	}

	//获取所有实例
	instancesInfo, err := client.Agent().Services()
	if err != nil {
		return nil, nil, nil, e.New(e.ConsulConnError, err)
	}

	for _, info := range instancesInfo {
		serviceList = append(serviceList, info.Service)
		IdInfo = append(IdInfo, *info)
	}

	//获取实例状态
	instancesStatus, err := client.Agent().Checks()
	if err != nil {
		return nil, nil, nil, e.New(e.ConsulConnError, err)
	}

	for _, info := range instancesStatus {
		serviceStatus[info.ServiceID] = *info
	}

	return IdInfo, serviceStatus, serviceList, nil
}
