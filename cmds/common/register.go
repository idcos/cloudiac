package common

import (
	"cloudiac/configs"
	"cloudiac/utils/consul"
	"cloudiac/utils/logs"
	"os"
)

func ServiceRegister(serviceName string) {
	conf := configs.Get()
	logger := logs.Get()

	logger.Debugf("Start register %s service", serviceName)
	err := consul.Register(serviceName, conf.Consul)
	if err != nil {
		logger.Debug("Service register failied: %s", err)
	} else {
		logger.Debug("Service register success.")
	}
}

func ReRegisterService(register bool, serviceName string) {
	ServiceRegister(serviceName)
	if register {
		os.Exit(0)
	}
}
