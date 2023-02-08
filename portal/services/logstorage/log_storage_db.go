// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package logstorage

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"os"
)

type dBLogStorage struct {
	db *db.Session
}

func (s *dBLogStorage) Write(path string, content []byte) error {
	_, err := s.db.Exec("REPLACE INTO iac_storage(path,content,created_at) VALUES (?,?,NOW())", path, content)
	return err
}

func (s *dBLogStorage) Read(path string) ([]byte, error) {
	dbLog := models.DBStorage{}
	if err := s.db.Where("path = ?", path).First(&dbLog); err != nil {
		if e.IsRecordNotFound(err) {
			return nil, os.ErrNotExist
		}
		return nil, err
	}
	return dbLog.Content, nil
}
