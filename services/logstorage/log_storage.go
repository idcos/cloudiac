package logstorage

import (
	"cloudiac/libs/db"
	"sync"
)

type LogStorage interface {
	Write(path string, content []byte) error
	Read(path string) ([]byte, error)
}

var (
	logStorage LogStorage
	initOnce   = sync.Once{}
)

func Get() LogStorage {
	initOnce.Do(func() {
		if logStorage == nil {
			logStorage = &dBLogStorage{db: db.Get()}
		}
	})
	return logStorage
}
