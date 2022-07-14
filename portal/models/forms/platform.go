// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package forms

type PfStatForm struct {
	PageForm
	OrgIds string `form:"orgIds" json:"orgIds"` // 组织ID
}
