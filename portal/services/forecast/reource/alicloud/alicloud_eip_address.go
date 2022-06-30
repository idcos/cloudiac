// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package alicloud

import (
	"cloudiac/portal/services/forecast/schema"
)

type EipAddress struct {
	Address  string
	Region   string
	Provider string
}

func (a *EipAddress) BuildResource() *schema.Resource {
	p := make([]schema.PriceRequest, 0)
	p = append(p, schema.PriceRequest{
		Type:      "eip",
		Attribute: map[string]string{},
	})

	return &schema.Resource{
		Name:        a.Address,
		Provider:    a.Provider,
		Region:      a.Region,
		RequestData: p,
	}
}
