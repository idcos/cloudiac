package logstorage

import (
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
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
