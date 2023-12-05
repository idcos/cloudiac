package db

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func init() {
	drivers["mysql"] = func(dsn string) gorm.Dialector {
		return mysql.New(mysql.Config{
			DSN:               dsn,
			DefaultStringSize: 255,
		})
	}
}
