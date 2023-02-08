// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package alicloud

import (
	"cloudiac/portal/services/forecast/schema"
)

type SlbLoadBalancer struct {
	Address          string
	Region           string
	Provider         string
	LoadBalancerSpec string `json:"load_balancer_spec"`
}

func (a *SlbLoadBalancer) BuildResource() *schema.Resource {
	p := make([]schema.PriceRequest, 0)
	if a.LoadBalancerSpec != "" {
		p = append(p, schema.PriceRequest{
			Type: "slb",
			Attribute: map[string]string{
				"spec": a.LoadBalancerSpec,
			},
		})
	}

	return &schema.Resource{
		Name:        a.Address,
		Provider:    a.Provider,
		Region:      a.Region,
		RequestData: p,
	}
}
