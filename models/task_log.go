package models

import "time"

type TaskLog struct {
	BaseModel
	CreatedAt time.Time

	Path    string `gorm:"NOT NULL;UNIQUE"`
	Content []byte `gorm:"type:BLOB"` // BLOB 支持最大长度约 64K
}

func (TaskLog) TableName() string {
	return "iac_task_log"
}
