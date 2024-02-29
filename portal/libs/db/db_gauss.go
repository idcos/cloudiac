package db

import (
	"github.com/jiangliuhong/gorm-driver-opengauss"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"strings"
)

func init() {
	d := Driver{
		Dialect: func(dsn string) gorm.Dialector {
			return postgres.Open(dsn)
		},
		SQLEnhance: func(sql string) string {
			sql = strings.ReplaceAll(sql, "`", "\"")
			return sql
		},
		Namer: postgres.Namer{
			NamingStrategy: schema.NamingStrategy{
				SingularTable: true,
			},
		},
	}

	drivers["gauss"] = d
}
