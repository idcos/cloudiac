// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

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
	p := make([]*schema.PriceRequest, 0)

	if a.InstanceType != "" {
		p = append(p, &schema.PriceRequest{
			//Name:  "DBInstanceClass",
			//Value: fmt.Sprintf("DBInstanceClass:%s,EngineVersion:%s", a.InstanceType, a.EngineVersion),
		})
	}

	if a.InstanceStorage != 0 {
		priceRequest := &schema.PriceRequest{
			//Name:  "DBInstanceStorage",
			//Value: fmt.Sprintf("DBInstanceStorage:%d,DBInstanceStorageType:%s", a.InstanceStorage, a.DbInstanceStorageType),
		}

		if a.DbInstanceStorageType != "" {
			//priceRequest.Value = fmt.Sprintf("DBInstanceStorage:%d,DBInstanceStorageType:%s", a.InstanceStorage, a.DbInstanceStorageType)
		} else {
			//priceRequest.Value = fmt.Sprintf("DBInstanceStorage:%d", a.InstanceStorage)
		}

		p = append(p, priceRequest)

	}

	return &schema.Resource{
		Name:        a.Address,
		Provider:    a.Provider,
		//RequestData: p,
		//PriceCode:   "rds",
		//PriceType:   "bards",
	}
}
