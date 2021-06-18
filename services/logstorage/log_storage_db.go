package logstorage

import (
	"cloudiac/consts"
	"cloudiac/libs/db"
	"cloudiac/models"
	"fmt"
)

type dBLogStorage struct {
	db *db.Session
}

func (s *dBLogStorage) Write(path string, content []byte) error {
	_, err := s.db.Save(&models.TaskLog{
		Path:    path,
		Content: content,
	})
	return err
}

func (s *dBLogStorage) Read(path string) ([]byte, error) {
	dbLog := models.TaskLog{}
	if err := s.db.Where("path = ?", path).First(&dbLog); err != nil {
		return nil, err
	}
	return dbLog.Content, nil
}

func (s *dBLogStorage) ReadStateList(templateGuid string) ([]byte, error) {
	dbLog := models.TaskLog{}
	if err := s.db.Where("path like ?", fmt.Sprintf("%s%s%s", templateGuid, "%", consts.TerraformStateListName)).Order("created_at desc").First(&dbLog); err != nil {
		return nil, err
	}
	return dbLog.Content, nil
}
