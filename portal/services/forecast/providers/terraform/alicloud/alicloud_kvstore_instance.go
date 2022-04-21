// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package alicloud

import (
	"cloudiac/portal/services/forecast/reource/alicloud"
	"cloudiac/portal/services/forecast/schema"
)

func getKvStoreInstanceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "alicloud_kvstore_instance",
		Notes: []string{},
		RFunc: NewSlb,
	}
}

func NewKvStoreInstance(d *schema.ResourceData) *schema.Resource {
	region := d.Get("region").String()

	a := &alicloud.Slb{
		Address:       d.Address,
		Provider:      d.ProviderName,
		Region:        region,
		Specification: d.Get("specification").String(),
	}

	return a.BuildResource()

}
