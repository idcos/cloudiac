// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package resps

type RespTag struct {
	KeyId   string `json:"keyId" form:"keyId" `
	ValueId string `json:"valueId" form:"valueId" `
	Key     string `json:"key" form:"key" `
	Value   string `json:"value" form:"value" `
	Source  string `json:"source" form:"source" `
}
