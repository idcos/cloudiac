// Copyright 2021 CloudJ Company Limited. All rights reserved.

package forms

type ConsulTagUpdateForm struct {
	PageForm

	Tags      []string `json:"tags" form:"tags" `
	ServiceId string   `json:"serviceId" form:"serviceId" `
}
