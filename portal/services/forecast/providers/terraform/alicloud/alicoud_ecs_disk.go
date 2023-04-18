// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.
package alicloud

import (
	"cloudiac/portal/services/forecast/reource/alicloud"
	"cloudiac/portal/services/forecast/schema"
)

func getEcsDiskRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "alicloud_ecs_disk",
		Notes: []string{},
		RFunc: NewEcsDisk,
	}
}

func NewEcsDisk(d *schema.ResourceData) *schema.Resource {
	a := &alicloud.EcsDisk{
		Address:          d.Address,
		Provider:         d.ProviderName,
		Region:           d.Region,
		Size:             d.Get("size").Int(),
		Category:         d.Get("category").String(),
		PerformanceLevel: d.Get("performance_level").String(),
	}

	a.InitDefault()

	return a.BuildResource()
}
