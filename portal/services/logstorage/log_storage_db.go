// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package logstorage

import (
	"cloudiac/configs"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"github.com/jiangliuhong/gorm-driver-dm/dmr"
	"os"
)

type dBLogStorage struct {
	db *db.Session
}

func (s *dBLogStorage) Write(path string, content []byte) error {
	dbType := configs.Get().GetDbType()
	var sql string
	var c interface{}
	if dbType == "dameng" {
		sql = `MERGE INTO iac_storage s
		using ( select ? path ,? as content ,NOW() as created_at)t
		on (s.path = t.path)
		when matched then
		update set content=t.content,created_at=t.created_at
		when not matched then
		insert (path,content,created_at) VALUES (t.path,t.content,t.created_at)`
		c = dmr.NewBlob(content)
	} else if dbType == "gauss" {
		sql = `MERGE INTO iac_storage s
		using ( select ? path ,HEXTORAW(?) as content ,NOW() as created_at)t
		on (s.path = t.path)
		when matched then
		update set content=t.content,created_at=t.created_at
		when not matched then
		insert (path,content,created_at) VALUES (t.path,t.content,t.created_at)`
		c = content
	} else {
		sql = "REPLACE INTO iac_storage(path,content,created_at) VALUES (?,?,NOW())"
		c = content
	}
	_, err := s.db.Exec(sql, path, c)
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
