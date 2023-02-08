// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package forms

type ConsulTagUpdateForm struct {
	PageForm

	Tags      []string `json:"tags" form:"tags" `
	ServiceId string   `json:"serviceId" form:"serviceId" `
}
