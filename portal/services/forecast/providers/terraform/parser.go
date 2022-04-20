// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package terraform

import (
	"cloudiac/portal/consts"
	"cloudiac/portal/services/forecast/schema"
	"github.com/tidwall/gjson"
	"path"
)

func ParserPlanJson(b []byte) (createResource []*schema.Resource, deleteResource []*schema.Resource, updateBeforeResource []*schema.Resource) {
	registryMap := GetResourceRegistryMap()
	parsed := gjson.ParseBytes(b)
	for _, v := range parsed.Get("resource_changes").Array() {
		t := v.Get("type").String()
		providerName := path.Base(v.Get("provider_name").String())
		address := v.Get("address").String()

		actions := v.Get("change.actions").Array()
		if len(actions) < 0 {
			continue
		}

		if actions[0].String() == consts.TerraformActionCreate {
			createResource = BuildResource(createResource, registryMap, t, providerName, address, v.Get("change.after"))
			//createResource = append(createResource, BuildResource(createResource,registryMap, t, providerName, address, v.Get("after")))
		}

		if actions[0].String() == consts.TerraformActionDelete {
			deleteResource = BuildResource(deleteResource, registryMap, t, providerName, address, v.Get("change.before"))
			//deleteResource = append(deleteResource, BuildResource(registryMap, t, providerName, address, v.Get("before")))
		}

		if actions[0].String() == consts.TerraformActionUpdate {
			createResource = BuildResource(createResource, registryMap, t, providerName, address, v.Get("change.after"))
			updateBeforeResource = BuildResource(updateBeforeResource, registryMap, t, providerName, address, v.Get("change.before"))
			//createResource = append(createResource, BuildResource(registryMap, t, providerName, address, v.Get("after")))
			//updateBeforeResource = append(updateBeforeResource, BuildResource(registryMap, t, providerName, address, v.Get("before")))
		}
	}

	return
}

func BuildResource(resource []*schema.Resource, registryMap *ResourceRegistryMap, t, providerName, address string, rawValues gjson.Result) []*schema.Resource {
	resourceData := schema.NewResourceData(t, providerName, address, rawValues)

	if registryItem, ok := (*registryMap)[t]; ok {
		resource = append(resource, registryItem.RFunc(resourceData))
	}
	return resource
}
