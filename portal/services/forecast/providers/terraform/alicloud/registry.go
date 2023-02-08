// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package alicloud

import "cloudiac/portal/services/forecast/schema"

var ResourceRegistry []*schema.RegistryItem = []*schema.RegistryItem{
	getInstanceRegistryItem(),
	getDiskRegistryItem(),
	getEcsDiskRegistryItem(),
	getNatGatewayRegistryItem(),
	getDBInstanceRegistryItem(),
	getEipRegistryItem(),
	getEipAddressRegistryItem(),
	getSlbRegistryItem(),
	getSlbLoadBalancerRegistryItem(),
	getKvStoreInstanceRegistryItem(),
	getMongodbInstanceRegistryItem(),
}
