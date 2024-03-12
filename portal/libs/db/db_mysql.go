package db

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func init() {
	drivers["mysql"] = Driver{
		Dialect: func(dsn string) gorm.Dialector {
			return mysql.New(mysql.Config{
				DSN:               dsn,
				DefaultStringSize: 255,
			})
		},
		SQLEnhance: func(sql string) string {
			return sql
		},
		Namer: defaultNamingStrategy,
	}
}
