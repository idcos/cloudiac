// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package alicloud

import (
	"cloudiac/portal/services/forecast/reource/alicloud"
	"cloudiac/portal/services/forecast/schema"
)

func getEipAddressRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "alicloud_eip_address",
		Notes: []string{},
		RFunc: NewEipAddress,
	}
}

func NewEipAddress(d *schema.ResourceData) *schema.Resource {

	e := &alicloud.Eip{
		Address:  d.Address,
		Provider: d.ProviderName,
		Region:   d.Region,
	}
	return e.BuildResource()

}
