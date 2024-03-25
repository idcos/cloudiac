package models

import (
	"cloudiac/portal/libs/db"
)

type ErrorMapping struct {
	AutoUintIdModel
	Manufacturer string `json:"manufacturer" gorm:"NOT NULL"`
	Level        string `json:"level" gorm:"NOT NULL"`
	Type         string `json:"type" gorm:"NOT NULL"`
	ErrorCode    string `json:"errorCode" gorm:"type:varchar(255);NOT NULL"`
	ErrorMessage string `json:"errorMessage" gorm:"NOT NULL"`
	Desc         string `json:"desc" gorm:"type:text"`
}

func (ErrorMapping) TableName() string {
	return "error_mapping"
}

func (u ErrorMapping) Migrate(sess *db.Session) error {
	return u.AddUniqueIndex(sess, "unique__error__mapping", "error_code")
}
