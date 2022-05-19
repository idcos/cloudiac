// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package alicloud

import (
	"cloudiac/portal/services/forecast/schema"
)

type MongodbInstance struct {
	Address  string
	Region   string
	Provider string

	DBInstanceClass string `json:"db_instance_class"`
}

func (a *MongodbInstance) BuildResource() *schema.Resource {
	p := make([]schema.PriceRequest, 0)

	if a.DBInstanceClass != "" {
		p = append(p, schema.PriceRequest{
			Type: "mongodb",
			Attribute: map[string]string{
				"db_instance_class": a.DBInstanceClass,
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
