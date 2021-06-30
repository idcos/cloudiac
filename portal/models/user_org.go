package models

import "cloudiac/portal/libs/db"

type UserOrg struct {
	BaseModel

	UserId uint   `json:"userId" gorm:"not null;comment:'用户ID'"`
	OrgId  uint   `json:"orgId" gorm:"not null;comment:'组织ID'"`
	Role   string `json:"role" gorm:"type:enum('owner','member');default:'member'"`
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
