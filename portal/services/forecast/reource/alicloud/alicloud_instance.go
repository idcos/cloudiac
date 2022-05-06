// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package alicloud

import (
	"cloudiac/portal/services/forecast/schema"
	"fmt"
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
}

type DataDisks struct {
	Category string `json:"category"`
	Size     int64  `json:"size"`
}

func (a *Instance) BuildResource() *schema.Resource {
	p := make([]*schema.PriceRequest, 0)

	if a.InstanceType != "" {
		p = append(p, &schema.PriceRequest{
			Name:  "InstanceType",
			Value: fmt.Sprintf("InstanceType:%s", a.InstanceType),
		})
	}

	if a.SystemDiskSize != 0 && a.SystemDiskCategory != "" {
		p = append(p, &schema.PriceRequest{
			Name:  "SystemDisk",
			Value: fmt.Sprintf("SystemDisk.Category:%s,SystemDisk.Size:%d", a.SystemDiskCategory, a.SystemDiskSize),
		})
	}

	if len(a.DataDisks) > 0 {
		for _, v := range a.DataDisks {
			p = append(p, &schema.PriceRequest{
				Name:  "DataDisk",
				Value: fmt.Sprintf("DataDisk.Category:%s,DataDisk.Size:%d", v.Category, v.Size),
			})
		}
	}

	return &schema.Resource{
		Name:        a.Address,
		Provider:    a.Provider,
		RequestData: p,
		PriceCode:   "ecs",
	}
}
