package models

type ResourceDrift struct {
	AutoUintIdModel
	ResId       Id     `json:"resId" gorm:"size:32;not null"`
	DriftDetail string `json:"driftDetail" gorm:"type:text"`
	CreateAt *Time `json:"createAt" gorm:"type:datetime;comment:任务开始时间"` // 任务开始时间
	UpdateAt *Time `json:"updateAt" gorm:"type:datetime;comment:任务开始时间"` // 任务开始时间
}

func (ResourceDrift) TableName() string {
	return "iac_resource_drift"
}
