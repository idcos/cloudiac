package models

type Tag struct {
	BaseModel
	Key        string `json:"key" gorm:"not null"`
	Value      string `json:"value" gorm:"not null"`
	Source     string `json:"source" gorm:"not null"`
	ObjectId   Id     `json:"objectId" gorm:"not null"`
	ObjectType string `json:"objectType" gorm:"not null"`
}

func (Tag) TableName() string {
	return "iac_tag"
}

