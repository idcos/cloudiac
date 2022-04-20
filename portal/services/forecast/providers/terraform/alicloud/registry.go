package alicloud

import "cloudiac/portal/services/forecast/schema"

var ResourceRegistry []*schema.RegistryItem = []*schema.RegistryItem{
	getInstanceRegistryItem(),
}

