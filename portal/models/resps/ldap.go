// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package resps

type LdapOUsResp struct {
	Dn  string   `json:"dn"`
	OUs []string `json:"ous"`
}

type LdapUsersResp struct {
	Dn    string `json:"dn"`
	Uid   string `json:"uid"`
	Email string `json:"email"`
}

type LdapUserAuthResp struct {
	Id string `json:"id"`
}

type LdapOUAuthResp struct {
	Id string `json:"id"`
}
