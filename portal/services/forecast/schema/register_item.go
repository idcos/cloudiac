// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package schema

type ReferenceIDFunc func(d *ResourceData) []string

type RegistryItem struct {
	Name                string
	Notes               []string
	RFunc               ResourceFunc
	ReferenceAttributes []string
	CustomRefIDFunc     ReferenceIDFunc
	NoPrice             bool
}
