// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package alicloud

import (
	"cloudiac/portal/services/forecast/schema"
	"fmt"
)

type NatGateway struct {
	Address            string
	Region             string
	Provider           string
	NatType            string `json:"nat_type"`
	Spec               string `json:"spec"`
	NetworkType        string `json:"network_type"`
	InternetChargeType string `json:"internet_charge_type"`
}

func (n *NatGateway) BuildResource() *schema.Resource {
	p := make([]*schema.PriceRequest, 0)

	p = append(p, &schema.PriceRequest{
		Name:  "NatType",
		Value: fmt.Sprintf("NatType:%s", n.NatType),
	})

	if n.Spec != "" {
		if n.InternetChargeType == "PayBySpec" && n.NetworkType == "internet" {
			p = append(p, &schema.PriceRequest{
				Name:  "Spec",
				Value: fmt.Sprintf("Spec:%s", n.Spec),
			})
		}
	}

	return &schema.Resource{
		Name:        n.Address,
		Provider:    n.Provider,
		RequestData: p,
		PriceCode:   "nat_gw",
	}
}
