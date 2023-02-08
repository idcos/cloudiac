// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package schema

type ResourceFunc func(*ResourceData) *Resource

type Resource struct {
	Name        string
	Region      string `json:"region"`
	Provider    string
	RequestData []PriceRequest
}

type PriceRequest struct {
	Type      string            `json:"type"`
	Attribute map[string]string `json:"attribute"`
}
