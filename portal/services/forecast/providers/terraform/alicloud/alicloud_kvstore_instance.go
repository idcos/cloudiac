// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package alicloud

import (
	"cloudiac/portal/services/forecast/reource/alicloud"
	"cloudiac/portal/services/forecast/schema"
)

func getKvStoreInstanceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "alicloud_kvstore_instance",
		Notes: []string{},
		RFunc: NewKvStoreInstance,
	}
}

func NewKvStoreInstance(d *schema.ResourceData) *schema.Resource {
	a := &alicloud.KvStoreInstance{
		Address:       d.Address,
		Provider:      d.ProviderName,
		Region:        d.Region,
		InstanceClass: d.Get("instance_class").String(),
	}

	return a.BuildResource()
}
