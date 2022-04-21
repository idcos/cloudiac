package consulClient

import (
	"cloudiac/common"
	"cloudiac/configs"
	consulapi "github.com/hashicorp/consul/api"
)

func NewConsulClient() (*consulapi.Client, error) {
	conf := configs.Get()
	config := consulapi.DefaultConfig()
	config.Address = conf.Consul.Address
	if conf.Consul.ConsulAcl {
		config.Token = conf.Consul.ConsulAclToken
	}

	if conf.Consul.ConsulTls {
		config.TLSConfig.CAFile = conf.Consul.ConsulCertPath + common.ConsulCa
		config.TLSConfig.CertFile = conf.Consul.ConsulCertPath + common.ConsulCapem
		config.TLSConfig.KeyFile = conf.Consul.ConsulCertPath + common.ConsulCakey
	}

	client, err := consulapi.NewClient(config)
	return client, err
}
