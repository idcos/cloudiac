// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package forms

type SearchLdapUserForm struct {
	BaseForm

	Q     string `json:"q" form:"q"`         // email 模糊检索
	Count int    `json:"count" form:"count"` // 检索数量
}

type SearchLdapOUForm struct {
	NoPageSizeForm
	FilterProjectId string `json:"filterProjectId" form:"filterProjectId"` // 过滤project关联的OU
}

type AuthLdapUserForm struct {
	BaseForm

	Emails []string `json:"email" form:"email"` // 邮箱
	Role   string   `json:"role" form:"role"`   // 角色
}

type AuthLdapOUForm struct {
	BaseForm

	DN   []string `json:"dn" form:"dn"`     // 识别名
	Role string   `json:"role" form:"role"` // 角色
}

type DeleteLdapOUForm struct {
	BaseForm

	Id string `json:"id" form:"id"` // ldap ou id
}

type UpdateLdapOUForm struct {
	BaseForm

	Id   string `json:"id" form:"id"` // ldap ou id
	Role string `json:"role" form:"role"`
}

type AuthProjectLdapOUForm struct {
	BaseForm

	DN   []string `json:"dn" form:"dn"`     // 识别名
	Role string   `json:"role" form:"role"` // 角色
}
