// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package alicloud

import (
	"cloudiac/portal/services/forecast/schema"
)

type Eip struct {
	Address  string
	Region   string
	Provider string
}

func (e *Eip) BuildResource() *schema.Resource {

	return &schema.Resource{
		Name:     e.Address,
		Provider: e.Provider,
		Region:   e.Region,
	}
}
