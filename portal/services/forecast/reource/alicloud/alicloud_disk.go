// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package alicloud

import (
	"cloudiac/portal/services/forecast/schema"
	"fmt"
	"strconv"
)

type Disk struct {
	Address          string
	Region           string
	Provider         string
	Category         string `json:"category"`
	Size             int64  `json:"size"`
	PerformanceLevel string `json:"performance_level"`
}

func (a *Disk) BuildResource() *schema.Resource {
	p := make([]schema.PriceRequest, 0)

	if a.Size > 0 && a.Category != "" {
		attribute := map[string]string{
			"type": a.Category,
			"size": strconv.Itoa(int(a.Size)),
		}

		if a.Category == "cloud_essd" {
			attribute["type"] = fmt.Sprintf("%s_%s", a.Category, a.PerformanceLevel)
		}

		p = append(p, schema.PriceRequest{
			Type:      "disk",
			Attribute: attribute,
		})
	}

	return &schema.Resource{
		Name:        a.Address,
		Provider:    a.Provider,
		RequestData: p,
		Region:      a.Region,
	}
}

func (a *Disk) InitDefault() {
	if a.Category == "" {
		a.Category = diskDefaultCategory
	}

	if a.PerformanceLevel == "" {
		a.PerformanceLevel = diskDefaultPerformanceLevel
	}
}
