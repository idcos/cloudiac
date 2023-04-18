// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

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
	a := &alicloud.Instance{
		Address:                    d.Address,
		Provider:                   d.ProviderName,
		Region:                     d.Region,
		InstanceType:               d.Get("instance_type").String(),
		SystemDiskSize:             d.Get("system_disk_size").Int(),
		SystemDiskCategory:         d.Get("system_disk_category").String(),
		SystemDiskPerformanceLevel: d.Get("system_disk_performance_level").String(),
	}
	disk := make([]alicloud.DataDisks, 0)
	for _, v := range d.Get("data_disks").Array() {
		disk = append(disk, alicloud.DataDisks{
			Category:         v.Get("category").String(),
			Size:             v.Get("size").Int(),
			PerformanceLevel: d.Get("performance_level").String(),
		})
	}

	a.DataDisks = disk

	a.InitDefault()

	return a.BuildResource()

}
