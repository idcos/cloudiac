// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package consul

import (
	"cloudiac/configs"
	"cloudiac/portal/services"
	"cloudiac/utils/consulClient"

	"cloudiac/utils/logs"

	"encoding/json"
	"fmt"
	"strings"
	"time"

	consulapi "github.com/hashicorp/consul/api"
)

func Register(serviceName string, consulConfig configs.ConsulConfig) error {
	logger := logs.Get()
	client, err := consulClient.NewConsulClient()

	if err != nil {
		logger.Errorf("consul client error : ", err)
		return err
	}
	consulTags, _ := services.ConsulKVSearch(consulConfig.ServiceID)
	var tags []string
	if consulConfig.ServiceTags != "" {
		tags = strings.Split(consulConfig.ServiceTags, ";")
	}
	if consulTags != nil && consulTags.(string) != "" {
		tags = []string{}
		_ = json.Unmarshal([]byte(consulTags.(string)), &tags)
	}

	registration := new(consulapi.AgentServiceRegistration)
	registration.ID = consulConfig.ServiceID      // 服务节点的名称
	registration.Name = serviceName               // 服务名称
	registration.Port = consulConfig.ServicePort  // 服务端口
	registration.Tags = tags                      // tag，可以为空
	registration.Address = consulConfig.ServiceIP // 服务 IP

	checkPort := consulConfig.ServicePort
	registration.Check = &consulapi.AgentServiceCheck{ // 健康检查
		HTTP:                           fmt.Sprintf("http://%s:%d/api/v1%s", registration.Address, checkPort, "/check"),
		Timeout:                        consulConfig.Timeout,
		Interval:                       consulConfig.Interval,        // 健康检查间隔
		DeregisterCriticalServiceAfter: consulConfig.DeregisterAfter, //check失败后30秒删除本服务，注销时间，相当于过期时间
	}

	err = client.Agent().ServiceRegister(registration)
	if err != nil {
		logger.Errorf("register server error : ", err)
		return err
	}
	return nil
}

func GetLocker(key string, value []byte, address string) (*consulapi.Lock, error) {
	client, err := consulClient.NewConsulClient()
	if err != nil {
		return nil, err
	}

	return client.LockOpts(&consulapi.LockOptions{
		Key:          key,
		Value:        value,
		SessionTTL:   "10s",       // session 超时时间, 超时后锁会被自动锁放
		LockTryOnce:  false,       // 重复尝试，直到加锁成功
		LockWaitTime: time.Second, // 加锁 api 请求的等待时间
		LockDelay:    time.Second, // consul server 的加锁操作等待时间
	})
}
