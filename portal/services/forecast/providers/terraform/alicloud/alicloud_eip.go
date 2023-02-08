// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package alicloud

import (
	"cloudiac/portal/services/forecast/reource/alicloud"
	"cloudiac/portal/services/forecast/schema"
)

func getEipRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "alicloud_eip",
		Notes: []string{},
		RFunc: NewEip,
	}
}

func NewEip(d *schema.ResourceData) *schema.Resource {
	a := &alicloud.Eip{
		Address:  d.Address,
		Provider: d.ProviderName,
		Region:   d.Region,
	}

	return a.BuildResource()

}
