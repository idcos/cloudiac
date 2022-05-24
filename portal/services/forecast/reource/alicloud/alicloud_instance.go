// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package alicloud

import (
	"cloudiac/portal/services/forecast/schema"
	"fmt"
	"strconv"
)

type Instance struct {
	Address                 string
	Region                  string
	Provider                string
	AvailabilityZone        string      `json:"availability_zone"`
	DataDisks               []DataDisks `json:"data_disks"`
	ImageId                 string      `json:"image_id"`
	InstanceType            string      `json:"instance_type"`
	InternetMaxBandwidthOut int         `json:"internet_max_bandwidth_out"`
	IoOptimized             interface{} `json:"io_optimized"`
	SystemDiskCategory      string      `json:"system_disk_category"`
	SystemDiskSize          int64       `json:"system_disk_size"`
	PerformanceLevel        string      `json:"performance_level"`
}

type DataDisks struct {
	Category         string `json:"category"`
	Size             int64  `json:"size"`
	PerformanceLevel string `json:"performance_level"`
}

func (a *Instance) BuildResource() *schema.Resource {
	p := make([]schema.PriceRequest, 0)

	if a.InstanceType != "" {
		p = append(p, schema.PriceRequest{
			Type: "ecs",
			Attribute: map[string]string{
				"instanceId": a.InstanceType,
			},
		})
	}

	if a.SystemDiskSize != 0 && a.SystemDiskCategory != "" {
		attribute := map[string]string{
			"type": a.SystemDiskCategory,
			"size": strconv.Itoa(int(a.SystemDiskSize)),
		}

		if a.SystemDiskCategory == "cloud_essd" {
			attribute["type"] = fmt.Sprintf("%s_%s", a.SystemDiskCategory, a.PerformanceLevel)
		}

		p = append(p, schema.PriceRequest{
			Type:      "disk",
			Attribute: attribute,
		})
	}

	if len(a.DataDisks) > 0 {
		for _, v := range a.DataDisks {
			attribute := map[string]string{
				"type": v.Category,
				"size": strconv.Itoa(int(v.Size)),
			}

			if a.SystemDiskCategory == "cloud_essd" {
				attribute["type"] = fmt.Sprintf("%s_%s", v.Category, v.PerformanceLevel)
			}

			p = append(p, schema.PriceRequest{
				Type:      "disk",
				Attribute: attribute,
			})
		}
	}

	return &schema.Resource{
		Name:        a.Address,
		Provider:    a.Provider,
		RequestData: p,
		Region:      a.Region,
	}
}
