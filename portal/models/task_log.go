package models

import (
	"cloudiac/portal/libs/db"
	"time"
)

type TaskLog struct {
	Id        uint   `gorm:"primary_key" json:"-"`
	Path      string `gorm:"NOT NULL;UNIQUE"`
	Content   []byte `gorm:"type:MEDIUMBLOB"` // MEDIUMBLOB 支持最大长度约 16M
	CreatedAt time.Time
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

func (TaskLog) Validate() error {
	return nil
}

func (TaskLog) ValidateAttrs(attrs Attrs) error {
	return nil
}
