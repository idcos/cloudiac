// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package alicloud

import (
	"cloudiac/portal/services/forecast/schema"
	"fmt"
)

type SlbLoadBalancer struct {
	Address  string
	Region   string
	Provider string

	Specification    string `json:"specification"`
	LoadBalancerSpec string `json:"load_balancer_spec"`
}

func (a *SlbLoadBalancer) BuildResource() *schema.Resource {
	p := make([]*schema.PriceRequest, 0)

	priceRequest := &schema.PriceRequest{
		Name: "LoadBalancerSpec",
	}
	if a.LoadBalancerSpec != "" {
		priceRequest.Value = fmt.Sprintf("LoadBalancerSpec:%s", a.LoadBalancerSpec)
	} else if a.Specification != "" {
		priceRequest.Value = fmt.Sprintf("LoadBalancerSpec:%s", a.Specification)
	}

	p = append(p, priceRequest)

	return &schema.Resource{
		Name:        a.Address,
		Provider:    a.Provider,
		RequestData: p,
		PriceCode:   "slb",
		PriceType:   "",
	}
}
