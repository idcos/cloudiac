// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package models

type LdapOUOrg struct {
	TimedModel

	OrgId Id     `json:"orgId" gorm:"size:32;not null;comment:组织ID"`                                   // 组织ID
	Role  string `json:"role" gorm:"type:enum('admin','complianceManager','member');default:'member'"` // 角色
	DN    string `json:"dn" gorm:"type:text"`                                                          // 识别名
	OU    string `json:"ou" gorm:"type:text"`                                                          // org units
}

func (LdapOUOrg) TableName() string {
	return "iac_ldap_ou_org"
}

type LdapOUProject struct {
	TimedModel

	ProjectId Id     `json:"projectId" gorm:"size:32;not null"`
	Role      string `json:"role" gorm:"type:enum('admin','complianceManager','member');default:'member'"` // 角色
	DN        string `json:"dn" gorm:"type:text"`                                                          // 识别名
	OU        string `json:"ou" gorm:"type:text"`                                                          // org units
}

func (LdapOUProject) TableName() string {
	return "iac_ldap_ou_project"
}
