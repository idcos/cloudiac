// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package alicloud

import (
	"cloudiac/portal/services/forecast/schema"
	"fmt"
	"strconv"
)

const (
	performanceLevel = "PL1"
	category         = "cloud_efficiency"
)

type Instance struct {
	Address                    string
	Region                     string
	Provider                   string
	AvailabilityZone           string      `json:"availability_zone"`
	DataDisks                  []DataDisks `json:"data_disks"`
	ImageId                    string      `json:"image_id"`
	InstanceType               string      `json:"instance_type"`
	InternetMaxBandwidthOut    int         `json:"internet_max_bandwidth_out"`
	IoOptimized                interface{} `json:"io_optimized"`
	SystemDiskCategory         string      `json:"system_disk_category"`
	SystemDiskSize             int64       `json:"system_disk_size"`
	SystemDiskPerformanceLevel string      `json:"system_disk_performance_level"`
}

type DataDisks struct {
	Category         string `json:"category"`
	Size             int64  `json:"size"`
	PerformanceLevel string `json:"performance_level"`
}

func (a *Instance) BuildResource() *schema.Resource {
	p := make([]schema.PriceRequest, 0)

	f := func(category, size, performanceLevel string) map[string]string {
		attribute := map[string]string{
			"type": category,
			"size": size,
		}
		if category == "cloud_essd" {
			attribute["type"] = fmt.Sprintf("%s_%s", category, performanceLevel)
		}

		return attribute
	}

	if a.InstanceType != "" {
		p = append(p, schema.PriceRequest{
			Type: "ecs",
			Attribute: map[string]string{
				"instanceId": a.InstanceType,
			},
		})
	}

	if a.SystemDiskSize != 0 && a.SystemDiskCategory != "" {
		p = append(p, schema.PriceRequest{
			Type:      "disk",
			Attribute: f(a.SystemDiskCategory, strconv.Itoa(int(a.SystemDiskSize)), a.SystemDiskPerformanceLevel),
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
				Attribute: f(v.Category, strconv.Itoa(int(v.Size)), v.PerformanceLevel),
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

func (a *Instance) InitDefault() {
	if a.SystemDiskCategory == "" {
		a.SystemDiskCategory = category
	}

	if a.SystemDiskPerformanceLevel == "" {
		a.SystemDiskPerformanceLevel = performanceLevel
	}

	for _, disk := range a.DataDisks {
		if disk.Category == "" {
			disk.Category = category
		}

		if disk.PerformanceLevel == "" {
			disk.PerformanceLevel = performanceLevel
		}
	}
}
