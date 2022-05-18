// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package alicloud

import (
	"cloudiac/portal/services/forecast/schema"
)

type Slb struct {
	Address  string
	Region   string
	Provider string

	Specification string `json:"specification"`
}

func (a *Slb) BuildResource() *schema.Resource {
	p := make([]*schema.PriceRequest, 0)

	if a.Specification != "" {
		p = append(p, &schema.PriceRequest{
			//Name:  "LoadBalancerSpec",
			//Value: fmt.Sprintf("LoadBalancerSpec:%s", a.Specification),
		})
	}

	return &schema.Resource{
		Name:        a.Address,
		Provider:    a.Provider,
		//RequestData: p,
		//PriceCode:   "slb",
		//PriceType:   "",
	}
}
