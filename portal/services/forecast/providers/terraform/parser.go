// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package terraform

import (
	"cloudiac/portal/services/forecast/schema"
	"github.com/tidwall/gjson"
	"path"
)

func ParserPlanJson(b []byte) []*schema.Resource {
	registryMap := GetResourceRegistryMap()
	resource := make([]*schema.Resource, 0)
	parsed := gjson.ParseBytes(b)
	for _, v := range parsed.Get("planned_values.root_module.resources").Array() {
		resourceData := schema.NewResourceData(v.Get("type").String(),
			path.Base(v.Get("provider_name").String()), v.Get("address").String(), v.Get("values"))

		if registryItem, ok := (*registryMap)[v.Get("type").String()]; ok {
			res := registryItem.RFunc(resourceData)
			resource = append(resource, res)
		}
	}

	return resource
}
