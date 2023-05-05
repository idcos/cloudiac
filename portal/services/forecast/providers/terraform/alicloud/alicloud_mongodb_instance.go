// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package alicloud

import (
	"cloudiac/portal/services/forecast/reource/alicloud"
	"cloudiac/portal/services/forecast/schema"
)

func getMongodbInstanceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "alicloud_mongodb_instance",
		Notes: []string{},
		RFunc: NewMongodbInstance,
	}
}

func NewMongodbInstance(d *schema.ResourceData) *schema.Resource {
	a := &alicloud.MongodbInstance{
		Address:         d.Address,
		Provider:        d.ProviderName,
		Region:          d.Region,
		DBInstanceClass: d.Get("db_instance_class").String(),
	}

	return a.BuildResource()
}
