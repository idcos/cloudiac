// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package resps

type LdapOUResp struct {
	DN       string       `json:"dn"`
	OU       string       `json:"ou"`
	Children []LdapOUResp `json:"children"`
}

type LdapOUDBResp struct {
	Id        string `json:"id"`
	DN        string `json:"dn"`
	OU        string `json:"ou"`
	Role      string `json:"role"`
	CreatedAt string `json:"createdAt"`
}

type LdapUserResp struct {
	DN          string `json:"dn"`
	Uid         string `json:"uid"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
}

type LdapUserListResp struct {
	LdapUsers []LdapUserResp `json:"ldapUsers"`
}

type AuthLdapUserResp struct {
	Ids []string `json:"ids"`
}

type AuthLdapOUResp struct {
	Ids []string `json:"ids"`
}

type DeleteLdapOUResp struct {
	Id string `json:"id"`
}

type UpdateLdapOUResp struct {
	Id string `json:"id"`
}

type OrgLdapOUsResp struct {
	DN string `json:"dn"`
	OU string `json:"ou"`
}

type OrgLdapOUListResp struct {
	OrgLdapOUs []OrgLdapOUsResp `json:"orgLdapOUs"`
}
