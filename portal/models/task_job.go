package models

type TaskJobBody struct {
}

type TaskJob struct {
	BaseModel

	TaskId      Id     `gorm:"size:32;not null"`
	Type        string `gorm:"type:enum('plan','apply','destroy','scan','parse')"`
	Image       string `gorm:""` // docker 镜像
	ContainerId string `gorm:"size:64"`
}

func (TaskJob) TableName() string {
	return "iac_task_job"
}
