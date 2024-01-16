package db

import (
	"github.com/jiangliuhong/gorm-driver-opengauss"
	"gorm.io/gorm"
)

func init() {
	d := Driver{
		Dialect: func(dsn string) gorm.Dialector {
			return postgres.Open(dsn)
		},
		SQLEnhance: func(sql string) string {
			return sql
		},
		Namer: defaultNamingStrategy,
	}

	drivers["gauss"] = d
}
