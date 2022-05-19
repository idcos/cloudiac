// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package alicloud

import (
	"cloudiac/portal/services/forecast/schema"
)

type NatGateway struct {
	Address  string
	Region   string
	Provider string
}

func (n *NatGateway) BuildResource() *schema.Resource {

	return &schema.Resource{
		Name:     n.Address,
		Provider: n.Provider,
		Region:   n.Region,
	}
}
