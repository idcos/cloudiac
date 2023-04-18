// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package terraform

import (
	"cloudiac/portal/services/forecast/providers/terraform/alicloud"
	"cloudiac/portal/services/forecast/schema"
)

type ResourceRegistryMap map[string]*schema.RegistryItem

var (
	resourceRegistryMap ResourceRegistryMap
)

func GetResourceRegistryMap() *ResourceRegistryMap {
	{
		resourceRegistryMap = make(ResourceRegistryMap)

		for _, registryItem := range alicloud.ResourceRegistry {
			resourceRegistryMap[registryItem.Name] = registryItem
		}
	}

	return &resourceRegistryMap
}
