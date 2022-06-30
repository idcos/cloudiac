// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package terraform

import (
	"cloudiac/portal/consts"
	"cloudiac/portal/services/forecast/schema"
	"fmt"
	"github.com/tidwall/gjson"
	"path"
	"strings"
)

const DEFAULT_ALICLOUD_REGION = "cn-beijing"

func ParserPlanJson(b []byte, aliRegion string) (createResource, deleteResource, updateBeforeResource, updateAfterResource []*schema.Resource) {
	registryMap := GetResourceRegistryMap()
	parsed := gjson.ParseBytes(b)

	// 查询阿里云的region属性
	aliConfig := parsed.Get("configuration.provider_config.alicloud.expressions.region")
	if aliConfig.Get("constant_value").Exists() {
		aliRegion = aliConfig.Get("constant_value").String()
	} else {
		references := aliConfig.Get("references").Array()
		for _, i := range references {
			if parsed.Get(fmt.Sprintf("variables.%s.value", strings.Replace(i.String(), "var.", "", 1))).Exists() {
				aliRegion = parsed.Get(fmt.Sprintf("variables.%s.value", strings.Replace(i.String(), "var.", "", 1))).String()
				break
			}
		}
	}
	if aliRegion == "" {
		aliRegion = DEFAULT_ALICLOUD_REGION
	}

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
				createResource = BuildResource(createResource, registryMap, t, providerName, address, aliRegion, v.Get("change.after"))
			}

			if action.String() == consts.TerraformActionDelete {
				deleteResource = BuildResource(deleteResource, registryMap, t, providerName, address, aliRegion, v.Get("change.before"))
			}

			if action.String() == consts.TerraformActionUpdate {
				updateAfterResource = BuildResource(updateAfterResource, registryMap, t, providerName, address, aliRegion, v.Get("change.after"))
				updateBeforeResource = BuildResource(updateBeforeResource, registryMap, t, providerName, address, aliRegion, v.Get("change.before"))
			}
		}
	}

	return
}

func BuildResource(resource []*schema.Resource, registryMap *ResourceRegistryMap, t, providerName, address, region string, rawValues gjson.Result) []*schema.Resource {
	resourceData := schema.NewResourceData(t, providerName, address, region, rawValues)

	if registryItem, ok := (*registryMap)[t]; ok {
		resource = append(resource, registryItem.RFunc(resourceData))
	}
	return resource
}
