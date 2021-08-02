package models

import (
	"cloudiac/portal/libs/db"
)

type DBStorage struct {
	AbstractModel

	Id        uint   `gorm:"primary_key" json:"-"`
	Path      string `gorm:"NOT NULL;UNIQUE"`
	Content   []byte `gorm:"type:MEDIUMBLOB"` // MEDIUMBLOB 支持最大长度约 16M
	CreatedAt Time   `gorm:"type:datetime"`
}

func (DBStorage) TableName() string {
	return "iac_storage"
}

func (DBStorage) Migrate(s *db.Session) error {
	if err := s.ModifyColumn(DBStorage{}.TableName(), "content"); err != nil {
		return err
	}
	return nil
}
