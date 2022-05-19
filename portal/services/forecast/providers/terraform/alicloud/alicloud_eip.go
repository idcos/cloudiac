// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

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
	region := d.Get("region").String()

	e := &alicloud.Eip{
		Address:  d.Address,
		Provider: d.ProviderName,
		Region:   region,
	}
	return e.BuildResource()

}
