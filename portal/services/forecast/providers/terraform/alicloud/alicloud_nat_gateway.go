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
	region := d.Get("region").String()

	n := &alicloud.NatGateway{
		Address:            d.Address,
		Provider:           d.ProviderName,
		Region:             region,
		NatType:            d.Get("nat_type").String(),
		Spec:               d.Get("specification").String(),
		NetworkType:        d.Get("network_type").String(),
		InternetChargeType: d.Get("internet_charge_type").String(),
	}
	return n.BuildResource()

}
