package models

type ResourceDrift struct {
	AutoUintIdModel
	EnvId          Id     `json:"envId" gorm:"size:32;not null"`
	CreateAt       *Time  `json:"createAt" gorm:"type:datetime"`
	TaskId         Id     `json:"taskId" gorm:"size:32;not null"`
	Address        string `json:"address" gorm:"size:32;not null"`
	ResourceDetail []byte `json:"resourceDetail" gorm:"type:MEDIUMBLOB"`
}

func (ResourceDrift) TableName() string {
	return "iac_resource_drift"
}
