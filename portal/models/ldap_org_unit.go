// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package models

type LdapOUOrg struct {
	TimedModel

	OrgId Id     `json:"orgId" gorm:"size:32;not null;comment:组织ID"` // 组织ID
	Role  string `json:"role" gorm:"default:'member'"`               // 角色 type:enum('admin','complianceManager','member')
	DN    Text   `json:"dn" gorm:"type:text"`                        // 识别名
	OU    Text   `json:"ou" gorm:"type:text"`                        // org units
}

func (LdapOUOrg) TableName() string {
	return "iac_ldap_ou_org"
}

type LdapOUProject struct {
	TimedModel

	OrgId     Id     `json:"orgId" gorm:"size:32;not null;comment:组织ID"` // 组织ID
	ProjectId Id     `json:"projectId" gorm:"size:32;not null"`
	Role      string `json:"role" gorm:"default:'operator';comment:角色"` // type:enum('manager','approver','operator','guest');
	DN        Text   `json:"dn" gorm:"type:text"`                       // 识别名
	OU        Text   `json:"ou" gorm:"type:text"`                       // org units
}

func (LdapOUProject) TableName() string {
	return "iac_ldap_ou_project"
}
