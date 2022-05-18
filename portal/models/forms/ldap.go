// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package forms

type SearchLdapUserForm struct {
	BaseForm

	Q     string `json:"q" form:"q"`         // email 模糊检索
	Count int    `json:"count" form:"count"` // 检索数量
}

type SearchLdapOUForm struct {
	NoPageSizeForm
}

type AuthLdapUserForm struct {
	BaseForm

	Email string `json:"email" form:"email"` // 邮箱
	Uid   string `json:"uid" form:"uid"`     // 用户名称
	Role  string `json:"role" form:"role"`   // 角色
	Phone string `json:"phone" form:"phone"` // 手机号
}

type AuthLdapOUForm struct {
	BaseForm

	DN   string `json:"dn" form:"dn"`     // 识别名
	OU   string `json:"ou" form:"ou"`     // organizationalUnit
	Role string `json:"role" form:"role"` // 角色
}

type DeleteLdapOUForm struct {
	BaseForm

	Id string `json:"id" form:"id"` // ldap ou id
}
