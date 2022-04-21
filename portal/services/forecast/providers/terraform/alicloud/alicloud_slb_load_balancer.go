// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package alicloud

import (
	"cloudiac/portal/services/forecast/reource/alicloud"
	"cloudiac/portal/services/forecast/schema"
)

func getSlbLoadBalancerRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "alicloud_slb_load_balancer",
		Notes: []string{},
		RFunc: NewSlbLoadBalancer,
	}
}

func NewSlbLoadBalancer(d *schema.ResourceData) *schema.Resource {
	region := d.Get("region").String()

	a := &alicloud.SlbLoadBalancer{
		Address:          d.Address,
		Provider:         d.ProviderName,
		Region:           region,
		Specification:    d.Get("specification").String(),
		LoadBalancerSpec: d.Get("load_balancer_spec").String(),
	}

	return a.BuildResource()

}
