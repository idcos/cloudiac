package models

type TaskComment struct {
	SoftDeleteModel
	TaskId    uint   `json:"taskId" form:"taskId" gorm:"not null;comment:'任务id'"`
	Creator   string `json:"creator" form:"creator" gorm:"not null;comment:'评论人'"`
	CreatorId uint   `json:"creatorId" form:"creatorId" gorm:"not null;comment:'评论人id'"`
	Comment   string `json:"comment" form:"comment"  gorm:"not null;comment:'评论'"`
}

func (TaskComment) TableName() string {
	return "iac_task_comment"
}
