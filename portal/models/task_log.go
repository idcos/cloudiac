package models

import (
	"cloudiac/portal/libs/db"
	"time"
)

type TaskLog struct {
	BaseModel
	CreatedAt time.Time

	Path    string `gorm:"NOT NULL;UNIQUE"`
	Content []byte `gorm:"type:MEDIUMBLOB"` // MEDIUMBLOB 支持最大长度约 16M
}

func (TaskLog) TableName() string {
	return "iac_task_log"
}

func (TaskLog) Migrate(s *db.Session) error {
	if err := s.DB().ModifyColumn("content", "MEDIUMBLOB").Error; err != nil {
		return err
	}
	return nil
}
