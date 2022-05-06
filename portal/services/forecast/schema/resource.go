// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package schema

type ResourceFunc func(*ResourceData) *Resource

type Resource struct {
	Name        string
	PriceType   string
	PriceCode   string
	Provider    string
	RequestData []*PriceRequest
}

type PriceRequest struct {
	Name  string
	Value string
}
