// Copyright 2021 CloudJ Company Limited. All rights reserved.

/*
独立于 model.Migrate() 的另一套 migrations 机制。
与 Migrate() 方法不同，该机制有版本控制，只在版本变化时执行一次 migrate func
*/

package migrations

import (
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/portal/services"
	"cloudiac/utils/logs"
	"fmt"
)

const (
	dbMigrateLock     = "cloudiac/db-migration-lock"
	dbMigrateLockWait = 60
)

func Run() error {
	logger := logs.Get().WithField("action", "migrate")

	dbSess := db.Get()
	var locked int
	if err := dbSess.Raw("SELECT GET_LOCK(?, ?) AS locked",
		dbMigrateLock, dbMigrateLockWait).Row().Scan(&locked); err != nil {
		return err
	}
	if locked == 0 {
		// 当有多个 portal 实例时只有一个会优先获得 lock 并执行 migrate，其他实例会等待其结束，
		// (相当于每一个实例都会判断一次是否需要执行 migrate，但同一时间只有一个实例在执行 migrate)。
		// migrate 操作应该在指定时间内执行完成，否则其他实例就会等待超时而报错，
		// 这样可以避免异常情况下锁没有被释放，所有实例都跳过了 migrate 执行，而程序未报错。
		logger.Panicf("GET_LOCK '%s' timeout", dbMigrateLock)
		return nil
	}

	defer func() {
		if _, err := dbSess.Exec("SELECT RELEASE_LOCK(?)", dbMigrateLock); err != nil {
			panic(err)
		}
	}()

	return db.Get().Transaction(func(tx *db.Session) error {
		var dbMigrationVersion = ""
		if tx.GormDB().Migrator().HasTable(models.SystemCfg{}.TableName()) {
			var err error
			dbMigrationVersion, err = services.GetMigrationVersion(tx)
			if err != nil {
				return err
			}
			logger.Infof("database migration version: '%s'", dbMigrationVersion)
		}

		dbVerIndex := -1 // 记录当前 db 数据库版本的索引
		verSet := make(map[string]struct{}, len(migrations))
		// 检查版本是否有重复
		for i, m := range migrations {
			if _, ok := verSet[m.Version]; ok {
				return fmt.Errorf("duplicate migration version: '%s'", m.Version)
			} else {
				verSet[m.Version] = struct{}{}
			}

			// 同时获取当前 db 中记录的 migration 版本在当前 migrations 列表中的索引
			if m.Version == dbMigrationVersion {
				dbVerIndex = i
			}
		}

		version := ""
		for _, m := range migrations[dbVerIndex+1:] {
			logger.Infof("run migrate '%s'", m.Version)
			for _, f := range m.Actions {
				if err := f(tx); err != nil {
					return err
				}
			}
			version = m.Version
		}

		if version != "" && version != dbMigrationVersion {
			logger.Infof("save migration version: '%s'", version)
			return services.SaveMigrationVersion(tx, version)
		}
		return nil
	})
}
