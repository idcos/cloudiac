// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package models

type LdapOrgUnit struct {
	BaseModel

	OrgId Id     `json:"orgId" gorm:"size:32;not null;comment:组织ID"`                                   // 组织ID
	Role  string `json:"role" gorm:"type:enum('admin','complianceManager','member');default:'member'"` // 角色
	DN    string `json:"dn" gorm:"type:text"`                                                          // 识别名
	OU    string `json:"ou" gorm:"type:text"`                                                          // org units
}

func (LdapOrgUnit) TableName() string {
	return "iac_ldap_org_unit"
}
