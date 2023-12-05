package db

import (
	dameng "github.com/jiangliuhong/gorm-driver-dm"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func init() {
	f := func(dsn string) gorm.Dialector {
		return dameng.Open("dm://" + dsn)
	}
	drivers["dm"] = f
	drivers["dameng"] = f
	n := dameng.Namer{schema.NamingStrategy{SingularTable: true}}
	namingStrategies["dm"] = n
	namingStrategies["dameng"] = n
}
