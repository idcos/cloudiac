// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package alicloud

import "cloudiac/portal/services/forecast/schema"

var ResourceRegistry []*schema.RegistryItem = []*schema.RegistryItem{
	getInstanceRegistryItem(),
	getDBInstanceRegistryItem(),
	getSlbRegistryItem(),
	getSlbLoadBalancerRegistryItem(),
	getKvStoreInstanceRegistryItem(),
}

