package migrations

import (
	"cloudiac/portal/migrations/types"
)

func newMigration(ver string, migrates ...types.MigrateFunc) types.Migration {
	return types.Migration{
		Version: ver,
		Actions: migrates,
	}
}

// 注册 migration 及其版本号。
// 版本号不可重复。
// 版本号需要按顺充填写，新的 migration 只能添加到列表末尾
var migrations = []types.Migration{
	newMigration("v20210905-init", initMigrate, dropDeletedAtColumn, renameRepoIdAndAddr, initLastResTaskId),
}
