// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package logstorage

import (
	"cloudiac/configs"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"os"
)

type dBLogStorage struct {
	db *db.Session
}

func (s *dBLogStorage) Write(path string, content []byte) error {
	dbType := configs.Get().GetDbType()
	var sql string
	if dbType == "dameng" {
		sql = `MERGE INTO iac_storage s
		using ( select 'test' path ,? as content ,NOW() as created_at)t
		on (s.path = t.path)
		when matched then
		update set content=t.content,created_at=t.created_at
		when not matched then
		insert (path,content,created_at) VALUES (t.path,t.content,t.created_at)`
	} else {
		sql = "REPLACE INTO iac_storage(path,content,created_at) VALUES (?,?,NOW())"
	}
	_, err := s.db.Exec(sql, path, content)
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
