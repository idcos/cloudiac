// Copyright 2021 CloudJ Company Limited. All rights reserved.

package models

import (
	"cloudiac/portal/libs/db"
)

type DBStorage struct {
	Id        uint   `gorm:"primary_key" json:"-"`
	Path      string `gorm:"NOT NULL;UNIQUE"`
	Content   []byte `gorm:"type:MEDIUMBLOB"` // MEDIUMBLOB 支持最大长度约 16M
	CreatedAt Time   `gorm:"type:datetime"`
}

func (DBStorage) TableName() string {
	return "iac_storage"
}

func (DBStorage) Migrate(s *db.Session) error {
	if err := s.DB().ModifyColumn("content", "MEDIUMBLOB").Error; err != nil {
		return err
	}
	return nil
}

func (DBStorage) Validate() error {
	return nil
}

func (DBStorage) ValidateAttrs(attrs Attrs) error {
	return nil
}
