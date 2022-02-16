// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package logstorage

import (
	"cloudiac/portal/consts"
	"cloudiac/portal/libs/db"
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

// CutLogContent 判断内容日志长度是否超限，若超限则截断(保留最新内容)
func CutLogContent(content []byte) []byte {
	size := len(content)
	if size > consts.MaxLogContentSize {
		content = content[size-consts.MaxLogContentSize:]
	}
	return content
}
