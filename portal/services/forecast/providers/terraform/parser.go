// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package terraform

import (
	"cloudiac/portal/consts"
	"cloudiac/portal/services/forecast/schema"
	"github.com/tidwall/gjson"
	"path"
)

func ParserPlanJson(b []byte) (createResource, deleteResource, updateBeforeResource, updateAfterResource []*schema.Resource) {
	registryMap := GetResourceRegistryMap()
	parsed := gjson.ParseBytes(b)
	for _, v := range parsed.Get("resource_changes").Array() {
		t := v.Get("type").String()
		providerName := path.Base(v.Get("provider_name").String())
		address := v.Get("address").String()
		actions := v.Get("change.actions").Array()
		if len(actions) <= 0 {
			continue
		}

		for _, action := range actions {
			if action.String() == consts.TerraformActionCreate {
				createResource = BuildResource(createResource, registryMap, t, providerName, address, v.Get("change.after"))
			}

			if action.String() == consts.TerraformActionDelete {
				deleteResource = BuildResource(deleteResource, registryMap, t, providerName, address, v.Get("change.before"))
			}

			if action.String() == consts.TerraformActionUpdate {
				updateAfterResource = BuildResource(updateAfterResource, registryMap, t, providerName, address, v.Get("change.after"))
				updateBeforeResource = BuildResource(updateBeforeResource, registryMap, t, providerName, address, v.Get("change.before"))
			}
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
