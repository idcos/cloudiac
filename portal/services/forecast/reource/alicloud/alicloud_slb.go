// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

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
	p := make([]schema.PriceRequest, 0)

	if a.Specification != "" {
		p = append(p, schema.PriceRequest{
			Type: "slb",
			Attribute: map[string]string{
				"spec": a.Specification,
			},
		})
	}

	return &schema.Resource{
		Name:        a.Address,
		Provider:    a.Provider,
		RequestData: p,
		Region:      a.Region,
	}
}
