// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package alicloud

import (
	"cloudiac/portal/services/forecast/schema"
)

type DBInstance struct {
	Address               string
	Region                string
	Provider              string
	AvailabilityZone      string `json:"availability_zone"`
	DbInstanceStorageType string `json:"db_instance_storage_type"`
	InstanceStorage       int64  `json:"instance_storage"`
	InstanceType          string `json:"instance_type"`
	Engine                string `json:"engine"`
	EngineVersion         string `json:"engine_version"`
}

func (a *DBInstance) BuildResource() *schema.Resource {
	p := make([]schema.PriceRequest, 0)

	if a.InstanceType != "" {
		p = append(p, schema.PriceRequest{
			Type: "rds",
			Attribute: map[string]string{
				"instance_type": a.InstanceType,
			},
		})
	}

	return &schema.Resource{
		Name:        a.Address,
		Provider:    a.Provider,
		Region:      a.Region,
		RequestData: p,
	}
}
