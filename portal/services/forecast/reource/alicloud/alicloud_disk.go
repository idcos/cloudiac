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
	Size     int64  `json:"size"`
}

func (a *Disk) BuildResource() *schema.Resource {
	//p := make([]*schema.PriceRequest, 0)

	if a.Size > 0 && a.Category != "" {
		//p = append(p, &schema.PriceRequest{
		//	Name:  "DataDisk.Category",
		//	Value: fmt.Sprintf("DataDisk.Category:%s,DataDisk.Size:%d", a.Category, a.Size),
		//})
	}

	return &schema.Resource{
		Name:        a.Address,
		Provider:    a.Provider,
		//RequestData: p,
		//PriceCode:   "yundisk",
		//PriceType:   "",
	}
}
