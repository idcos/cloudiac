package migrations

import (
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"fmt"
)

type Migrator interface {
	Migrate(*db.Session) error
}

func initMigrate(tx *db.Session) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	autoMigrate := func(m interface{}, tx *db.Session) {
		tx = tx.Model(m)
		if err := tx.GormDB().AutoMigrate(m); err != nil {
			panic(fmt.Errorf("auto migrate %T: %v", m, err))
		}
		if mg, ok := m.(Migrator); ok {
			if err := mg.Migrate(tx); err != nil {
				panic(fmt.Errorf("auto migrate %T: %v", m, err))
			}
		}
	}

	// 以下全部为旧版本的 model Auto Migrate 调用，
	// 改用新的 migrate 机制后，新添加的 model 应该新创建一个 migration version

	autoMigrate(&models.Organization{}, tx)
	autoMigrate(&models.Project{}, tx)
	autoMigrate(&models.Vcs{}, tx)
	autoMigrate(&models.Template{}, tx)
	autoMigrate(&models.Env{}, tx)
	autoMigrate(&models.Resource{}, tx)

	autoMigrate(&models.Variable{}, tx)

	autoMigrate(&models.Task{}, tx)
	autoMigrate(&models.TaskStep{}, tx)
	autoMigrate(&models.DBStorage{}, tx)

	autoMigrate(&models.User{}, tx)
	autoMigrate(&models.UserOrg{}, tx)
	autoMigrate(&models.UserProject{}, tx)

	autoMigrate(&models.Notification{}, tx)
	autoMigrate(&models.NotificationEvent{}, tx)
	autoMigrate(&models.SystemCfg{}, tx)
	autoMigrate(&models.ResourceAccount{}, tx)
	autoMigrate(&models.CtResourceMap{}, tx)
	autoMigrate(&models.OperationLog{}, tx)
	autoMigrate(&models.Token{}, tx)
	autoMigrate(&models.Key{}, tx)
	autoMigrate(&models.TaskComment{}, tx)
	autoMigrate(&models.ProjectTemplate{}, tx)
	return nil
}

func dropDeletedAtColumn(tx *db.Session) error {
	for _, t := range []interface{}{
		&models.Env{},
		&models.User{},
		&models.Task{},
		&models.Project{},
		&models.Template{},
	} {
		if err := tx.DropColumn(t, "deleted_at"); err != nil {
			return err
		}
	}
	return nil
}

func renameRepoIdAndAddr(tx *db.Session) error {
	sqls := []string{
		"update iac_template SET repo_id = replace(repo_id,'/cloud-iac/','/cloudiac/') " +
			"WHERE repo_id like '/cloud-iac/%'",
		"update iac_template SET repo_addr = replace(repo_addr,'/repos/cloud-iac/','/repos/cloudiac/') " +
			"WHERE repo_addr like '%/repos/cloud-iac/%'",
	}
	for _, sql := range sqls {
		if _, err := tx.Exec(sql); err != nil {
			return err
		}
	}
	return nil
}

func initLastResTaskId(tx *db.Session) error {
	cnt, err := tx.Model(&models.Env{}).Where("last_res_task_id IS NOT NULL").Count()
	if err != nil {
		return err
	}
	if cnt > 0 {
		// 己经有记录的  last_res_task_id 不为 null，表示己经手动执行过下面的 update 操作，我们直接返回
		return nil
	}
	_, err = tx.Exec("UPDATE iac_env SET last_res_task_id=last_task_id WHERE last_res_task_id IS NULL")
	return err
}
