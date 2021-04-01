package common

import (
	"os"
	"cloudiac/configs"
	"cloudiac/utils/consul"
	"cloudiac/utils/logs"
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