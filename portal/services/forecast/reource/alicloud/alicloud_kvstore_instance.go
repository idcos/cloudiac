// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package alicloud

import (
	"cloudiac/portal/services/forecast/schema"
)

type KvStoreInstance struct {
	Address  string
	Region   string
	Provider string

	InstanceClass string `json:"instance_class"`
	InstanceType  string `json:"instance_type"`
}

func (a *KvStoreInstance) BuildResource() *schema.Resource {
	p := make([]schema.PriceRequest, 0)

	if a.InstanceClass != "" {
		p = append(p, schema.PriceRequest{
			Type: "redis",
			Attribute: map[string]string{
				"instance_class": a.InstanceClass,
			},
		})
	}

	return &schema.Resource{
		Name:        a.Address,
		Provider:    a.Provider,
		Region:      a.Region,
		RequestData: p,
	}
}
