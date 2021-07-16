package models

import "cloudiac/portal/libs/db"

type UserOrg struct {
	BaseModel

	UserId Id     `json:"userId" gorm:"size:32;not null;comment:'用户ID'"`            // 用户ID
	OrgId  Id     `json:"orgId" gorm:"size:32;not null;comment:'组织ID'"`             // 组织ID
	Role   string `json:"role" gorm:"type:enum('admin','member');default:'member'"` // 角色
}

func (UserOrg) TableName() string {
	return "iac_user_org"
}

func (m UserOrg) Migrate(sess *db.Session) (err error) {
	err = m.AddUniqueIndex(sess, "unique__org_id__user_id", "org_id", "user_id")
	if err != nil {
		return err
	}

	return nil
}
