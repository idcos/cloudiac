// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package resps

type LdapOUResp struct {
	Dn  string   `json:"dn"`
	OUs []string `json:"ous"`
}

type LdapOUListResp struct {
	LdapOUs []LdapOUResp `json:"ldap_ous"`
}

type LdapUserResp struct {
	Dn    string `json:"dn"`
	Uid   string `json:"uid"`
	Email string `json:"email"`
}

type LdapUserListResp struct {
	LdapUsers []LdapUserResp `json:"ldap_users"`
}

type LdapUserAuthResp struct {
	Id string `json:"id"`
}

type LdapOUAuthResp struct {
	Id string `json:"id"`
}
