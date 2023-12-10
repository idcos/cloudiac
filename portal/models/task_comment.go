// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package models

import "cloudiac/portal/libs/db"

type TaskComment struct {
	TimedModel
	TaskId    Id     `json:"taskId" form:"taskId" gorm:"size:32;not null;comment:任务id"`
	Creator   string `json:"creator" form:"creator" gorm:"size:32;not null;comment:评论人"`
	CreatorId Id     `json:"creatorId" form:"creatorId" gorm:"size:32;not null;comment:评论人id"`
	Comment   Text   `json:"comment" form:"comment"  gorm:"type:text;comment:评论"`
}

func (TaskComment) TableName() string {
	return "iac_task_comment"
}

func (TaskComment) Migrate(tx *db.Session) error {
	if err := tx.ModifyModelColumn(&TaskComment{}, "`comment`"); err != nil {
		return err
	}
	return nil
}
