// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package services

import (
	"cloudiac/common"
	"cloudiac/configs"
	"cloudiac/portal/consts/e"
	"cloudiac/utils"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/consul/api"
	"github.com/pkg/errors"
)

func SystemStatusSearch() ([]api.AgentService, map[string]api.AgentCheck, []string, e.Error) {
	serviceList := make([]string, 0)
	IdInfo := make([]api.AgentService, 0)
	serviceStatus := make(map[string]api.AgentCheck)
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

func SystemRunnerTags() ([]string, e.Error) {

	conf := configs.Get()
	config := api.DefaultConfig()
	config.Address = conf.Consul.Address
	tags := make([]string, 0)

	client, err := api.NewClient(config)
	if err != nil {
		return nil, e.New(e.ConsulConnError, err)
	}

	//获取所有实例
	instancesInfo, err := client.Agent().Services()
	if err != nil {
		return nil, e.New(e.ConsulConnError, err)
	}

	for _, info := range instancesInfo {
		if info.Service != common.RunnerServiceName {
			continue
		}
		tags = append(tags, info.Tags...)
	}

	return utils.RemoveDuplicateElement(tags), nil
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
	if value != nil && value.Value != nil {
		return string(value.Value), nil
	}
	return nil, nil

}

func RunnerSearch() ([]*api.AgentService, e.Error) {
	resp := make([]*api.AgentService, 0)

	conf := configs.Get()
	config := api.DefaultConfig()
	config.Address = conf.Consul.Address

	client, err := api.NewClient(config)
	if err != nil {
		return nil, e.New(e.ConsulConnError, err)
	}

	checkes, _, err := client.Health().Checks(
		common.RunnerServiceName,
		&api.QueryOptions{
			Filter: "Status==passing",
		},
	)
	if err != nil {
		return nil, e.New(e.ConsulConnError, err)
	}
	passingServices := make(map[string]struct{})
	for _, c := range checkes {
		passingServices[c.ServiceID] = struct{}{}
	}

	services, err := client.Agent().Services()
	if err != nil {
		return nil, e.New(e.ConsulConnError, err)
	}

	for _, s := range services {
		if _, ok := passingServices[s.ID]; ok && s.Service == common.RunnerServiceName {
			resp = append(resp, s)
		}
	}

	return resp, nil
}

func ConsulKVSave(key string, values []string) e.Error {
	conf := configs.Get()
	config := api.DefaultConfig()
	config.Address = conf.Consul.Address

	client, err := api.NewClient(config)
	if err != nil {
		return e.New(e.ConsulConnError, err)
	}
	b, _ := json.Marshal(values)
	_, err = client.KV().Put(&api.KVPair{Key: key, Value: []byte(b)}, nil)
	if err != nil {
		return e.New(e.ConsulConnError, err)
	}
	return nil

}

func ConsulServiceInfo(serviceId string) (*api.AgentService, e.Error) {
	conf := configs.Get()
	config := api.DefaultConfig()
	config.Address = conf.Consul.Address

	client, err := api.NewClient(config)
	if err != nil {
		return nil, e.New(e.ConsulConnError, err)
	}
	agentService, _, err := client.Agent().Service(serviceId, nil)
	if err != nil {
		return nil, e.New(e.ConsulConnError, err)
	}
	return agentService, nil

}

func ConsulServiceRegistered(serviceInfo *api.AgentService, tags []string) e.Error {
	consulConfig := configs.Get().Consul
	config := api.DefaultConfig()
	config.Address = consulConfig.Address
	client, err := api.NewClient(config)
	if err != nil {
		return e.New(e.ConsulConnError, fmt.Errorf("consul client error : %v", err))
	}

	registration := new(api.AgentServiceRegistration)
	registration.ID = serviceInfo.ID           // 服务节点的名称
	registration.Name = serviceInfo.Service    // 服务名称
	registration.Port = serviceInfo.Port       // 服务端口
	registration.Tags = tags                   // tag，可以为空
	registration.Address = serviceInfo.Address // 服务 IP

	checkPort := serviceInfo.Port
	registration.Check = &api.AgentServiceCheck{ // 健康检查
		HTTP:                           fmt.Sprintf("http://%s:%d/api/v1%s", registration.Address, checkPort, "/check"),
		Timeout:                        consulConfig.Timeout,
		Interval:                       consulConfig.Interval,        // 健康检查间隔
		DeregisterCriticalServiceAfter: consulConfig.DeregisterAfter, //check失败后30秒删除本服务，注销时间，相当于过期时间
	}

	err = client.Agent().ServiceRegister(registration)
	if err != nil {
		return e.New(e.ConsulConnError, fmt.Errorf("register server error : %v", err))
	}
	return nil
}

func GetRunnerAddress(serviceId string) (string, error) {
	s, err := ConsulServiceInfo(serviceId)
	if err != nil {
		return "", errors.Wrapf(err, "get runner address, runnerId %s", serviceId)
	}
	return fmt.Sprintf("http://%s:%d", s.Address, s.Port), nil
}

func GetDefaultRunnerId() (string, e.Error) {
	runners, err := RunnerSearch()
	if err != nil {
		return "", err
	}
	if len(runners) > 0 {
		return runners[0].ID, nil
	}
	return "", e.New(e.ConsulConnError, fmt.Errorf("no active runner found"))
}
