// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package alicloud

import (
	"cloudiac/portal/services/forecast/schema"
)

type Disk struct {
	Address  string
	Region   string
	Provider string

	Category string `json:"category"`
	Size     int    `json:"size"`
}

func (a *Disk) BuildResource() *schema.Resource {
	//p := make([]*schema.PriceRequest, 0)

	return &schema.Resource{
		Name:        a.Address,
		Provider:    a.Provider,
		//RequestData: p,
		//PriceCode:   "yundisk",
		//PriceType:   "",
	}
}
