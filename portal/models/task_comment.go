// Copyright 2021 CloudJ Company Limited. All rights reserved.

package models

type TaskComment struct {
	TimedModel
	TaskId    Id     `json:"taskId" form:"taskId" gorm:"size:32;not null;comment:任务id"`
	Creator   string `json:"creator" form:"creator" gorm:"size:32;not null;comment:评论人"`
	CreatorId Id     `json:"creatorId" form:"creatorId" gorm:"size:32;not null;comment:评论人id"`
	Comment   string `json:"comment" form:"comment"  gorm:"not null;comment:评论"`
}

func (TaskComment) TableName() string {
	return "iac_task_comment"
}
