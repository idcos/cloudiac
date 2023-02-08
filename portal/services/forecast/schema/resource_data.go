// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package schema

import "github.com/tidwall/gjson"

type ResourceData struct {
	Type          string
	ProviderName  string
	Address       string
	Region        string
	Tags          map[string]string
	RawValues     gjson.Result
	referencesMap map[string][]*ResourceData
}

func NewResourceData(resourceType, providerName, address, region string, rawValues gjson.Result) *ResourceData {
	return &ResourceData{
		Type:          resourceType,
		ProviderName:  providerName,
		Address:       address,
		Region:        region,
		RawValues:     rawValues,
		referencesMap: make(map[string][]*ResourceData),
	}
}

func (d *ResourceData) Get(key string) gjson.Result {
	return d.RawValues.Get(key)
}
