package models

import "cloudiac/portal/libs/db"

type UserProject struct {
	BaseModel

	UserId    uint   `json:"userId" gorm:"not null;comment:'用户ID'"`
	ProjectId uint   `json:"projectId" gorm:"not null"`
	Role      string `json:"role" gorm:"type:enum('owner','manager','operator','guest');default:'operator';comment:'角色'"`
}

func (UserProject) TableName() string {
	return "iac_user_project"
}

func (u UserProject) Migrate(sess *db.Session) error {
	return u.AddUniqueIndex(sess, "unique__user__project", "user_id", "project_id")
}
