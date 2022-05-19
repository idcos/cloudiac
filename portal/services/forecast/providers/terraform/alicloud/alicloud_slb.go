// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package alicloud

import (
	"cloudiac/portal/services/forecast/reource/alicloud"
	"cloudiac/portal/services/forecast/schema"
)

func getSlbRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "alicloud_slb",
		Notes: []string{},
		RFunc: NewSlb,
	}
}

func NewSlb(d *schema.ResourceData) *schema.Resource {

	a := &alicloud.Slb{
		Address:       d.Address,
		Provider:      d.ProviderName,
		Region:        d.Region,
		Specification: d.Get("specification").String(),
	}

	return a.BuildResource()

}
