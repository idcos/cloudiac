package models

type ResourceDrift struct {
	AutoUintIdModel
	ResId       Id     `json:"resId" gorm:"size:32;not null"`
	DriftDetail string `json:"driftDetail" gorm:"type:text"`
}

func (ResourceDrift) TableName() string {
	return "iac_resource_drift"
}
