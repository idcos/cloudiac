// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package models

import (
	"cloudiac/portal/libs/db"
)

type DBStorage struct {
	AbstractModel

	Id        uint     `gorm:"primaryKey" json:"-"`
	Path      string   `gorm:"NOT NULL;UNIQUE"`
	Content   ByteBlob `gorm:""` // LONGBLOB 支持最大长度约 4G 达梦不支持LONGBLOB，改为blob
	CreatedAt Time     `gorm:""`
}

func (DBStorage) TableName() string {
	return "iac_storage"
}

func (DBStorage) Migrate(s *db.Session) error {
	if err := s.ModifyModelColumn(&DBStorage{}, "content"); err != nil {
		return err
	}
	return nil
}
