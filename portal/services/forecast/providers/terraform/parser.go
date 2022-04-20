package terraform

import (
	"cloudiac/portal/services/forecast/schema"
	"github.com/tidwall/gjson"
)

func ParserPlanJson(b []byte) []*schema.Resource {
	registryMap := GetResourceRegistryMap()
	resource := make([]*schema.Resource, 0)
	parsed := gjson.ParseBytes(b)
	for _, v := range parsed.Get("planned_values.root_module.resources").Array() {
		r := &schema.ResourceData{
			Type:         v.Get("type").String(),
			RawValues:    v.Get("values"),
			Address:      v.Get("address").String(),
			ProviderName: v.Get("provider_name").String(),
		}

		if registryItem, ok := (*registryMap)[v.Get("type").String()]; ok {
			res := registryItem.RFunc(r)
			resource = append(resource, res)
		}
	}
	return resource
}
