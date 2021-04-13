package services

import (
	"cloudiac/configs"
	"cloudiac/consts/e"
	"errors"
	"github.com/hashicorp/consul/api"
	"strings"
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

func ConsulKVSearch(key string) (interface{}, e.Error) {
	conf := configs.Get()
	config := api.DefaultConfig()
	config.Address = conf.Consul.Address

	client, err := api.NewClient(config)
	if err != nil {
		return nil, e.New(e.ConsulConnError, err)
	}
	value, _, err := client.KV().Get(key, nil)
	if err != nil {
		return nil, e.New(e.ConsulConnError, err)
	}
	if value != nil && value.Value != nil{
		return string(value.Value), nil
	}
	return nil,e.New(e.BadParam,errors.New("key error"))

}

func RunnerListSearch() (interface{}, e.Error) {
	resp := make([]*api.AgentService, 0)

	conf := configs.Get()
	config := api.DefaultConfig()
	config.Address = conf.Consul.Address

	client, err := api.NewClient(config)
	if err != nil {
		return nil, e.New(e.ConsulConnError, err)
	}
	services, err := client.Agent().Services()
	if err != nil {
		return nil, e.New(e.ConsulConnError, err)
	}

	for serviceName, _ := range services {
		if strings.Contains(strings.ToLower(serviceName), "runner") {
			resp = append(resp, services[serviceName])
		}
	}

	return resp, nil
}
