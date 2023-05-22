// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package forms

type DeployForm struct {
	Units   map[string]Unit   `json:"units"`
	Outputs map[string]string `json:"outputs"`
}

type Unit struct {
	Module string              `json:"module"`
	Count  int                 `json:"count"`
	Vars   []map[string]string `json:"vars"`
}
