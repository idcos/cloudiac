// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package alicloud

import (
	"cloudiac/portal/services/forecast/reource/alicloud"
	"cloudiac/portal/services/forecast/schema"
)

func getInstanceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "alicloud_instance",
		Notes: []string{},
		RFunc: NewInstance,
	}
}

func NewInstance(d *schema.ResourceData) *schema.Resource {
	region := d.Get("region").String()

	a := &alicloud.Instance{
		Address:            d.Address,
		Provider:           d.ProviderName,
		Region:             region,
		InstanceType:       d.Get("instance_type").String(),
		SystemDiskSize:     d.Get("system_disk_size").Int(),
		SystemDiskCategory: d.Get("system_disk_category").String(),
	}
	disk := make([]alicloud.DataDisks, 0)
	for _, v := range d.Get("data_disks").Array() {
		disk = append(disk, alicloud.DataDisks{
			Category: v.Get("category").String(),
			Size:     v.Get("size").Int(),
		})
	}

	a.DataDisks = disk

	return a.BuildResource()

}
