package db

import (
	dameng "github.com/jiangliuhong/gorm-driver-dm"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"strings"
)

func init() {

	d := Driver{
		Dialect: func(dsn string) gorm.Dialector {
			return dameng.Open("dm://" + dsn)
		},
		SQLEnhance: func(sql string) string {
			// 替换 `` 符号
			sql = strings.ReplaceAll(sql, "`", "\"")
			return sql
		},
		Namer: dameng.Namer{schema.NamingStrategy{SingularTable: true}},
	}

	drivers["dameng"] = d
	drivers["dm"] = d
}
