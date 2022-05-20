// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package alicloud

import (
	"cloudiac/portal/services/forecast/schema"
	"strconv"
)

type EcsDisk struct {
	Address  string
	Region   string
	Provider string
	Category string `json:"category"`
	Size     int64  `json:"size"`
}

func (a *EcsDisk) BuildResource() *schema.Resource {
	p := make([]schema.PriceRequest, 0)

	if a.Size > 0 && a.Category != "" {
		p = append(p, schema.PriceRequest{
			Type: "disk",
			Attribute: map[string]string{
				"type": a.Category,
				"size": strconv.Itoa(int(a.Size)),
			},
		})
	}

	return &schema.Resource{
		Name:        a.Address,
		Provider:    a.Provider,
		RequestData: p,
		Region:      a.Region,
		//RequestData: p,
		//PriceCode:   "yundisk",
		//PriceType:   "",
	}
}
