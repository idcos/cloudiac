// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package alicloud

import (
	"cloudiac/portal/services/forecast/reource/alicloud"
	"cloudiac/portal/services/forecast/schema"
)

func getNatGatewayRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "alicloud_nat_gateway",
		Notes: []string{},
		RFunc: NewNatGateway,
	}
}

func NewNatGateway(d *schema.ResourceData) *schema.Resource {

	n := &alicloud.NatGateway{
		Address:  d.Address,
		Provider: d.ProviderName,
		Region:   d.Region,
	}
	return n.BuildResource()

}
