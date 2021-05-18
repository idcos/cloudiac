package consul

import (
	"cloudiac/configs"
	"cloudiac/services"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	consulapi "github.com/hashicorp/consul/api"
)

func Register(serviceName string, consulConfig configs.ConsulConfig) error {
	config := consulapi.DefaultConfig()
	config.Address = consulConfig.Address
	client, err := consulapi.NewClient(config)
	if err != nil {
		log.Fatal("consul client error : ", err)
		return err
	}
	consulTags, _ := services.ConsulKVSearch(serviceName)
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
		log.Fatal("register server error : ", err)
		return err
	}
	return nil
}
