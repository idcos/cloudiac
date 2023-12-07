// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package models

import (
	"cloudiac/configs"
	"cloudiac/utils/logs"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"github.com/jiangliuhong/gorm-driver-dm/dmr"
	dmSchema "github.com/jiangliuhong/gorm-driver-dm/schema"

	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
)

type errorFunc func(tx *db.Session) (int64, error)

func withTx(tx *db.Session, f errorFunc) (int64, error) {
	localTx := false
	if tx == nil {
		tx = db.Get().Begin()
		localTx = true
		defer func() {
			if r := recover(); r != nil {
				_ = tx.Rollback()
				panic(r)
			}
		}()
	}

	n, err := f(tx)
	if err != nil {
		if localTx {
			_ = tx.Rollback()
		}
		return n, err
	} else if localTx {
		return n, tx.Commit()
	}
	return n, nil
}

func Validate(tx *db.Session, o Modeler) error {
	return o.Validate()
}

func Create(tx *db.Session, o Modeler) error {
	_, err := withTx(tx, func(x *db.Session) (int64, error) {
		if err := o.Validate(); err != nil {
			return 0, err
		}
		if err := x.Insert(o); err != nil {
			return 0, err
		}
		return 0, nil
	})
	return err
}

// CreateBatch 注意: 目前切片 Modeler 类型无法与批量插入公用
func CreateBatch(tx *db.Session, o interface{}) error {
	_, err := withTx(tx, func(x *db.Session) (int64, error) {
		if err := x.Insert(o); err != nil {
			return 0, err
		}
		return 0, nil
	})
	return err
}

func Save(tx *db.Session, o Modeler) error {
	_, err := withTx(tx, func(x *db.Session) (int64, error) {
		if err := o.Validate(); err != nil {
			return 0, err
		}
		if _, err := x.Save(o); err != nil {
			return 0, err
		}
		return 0, nil
	})
	return err
}

func UpdateAttr(tx *db.Session, o Modeler, values Attrs, query ...interface{}) (int64, error) {
	return withTx(tx, func(x *db.Session) (int64, error) {
		if err := o.ValidateAttrs(values); err != nil {
			return 0, e.AutoNew(err, e.DBAttrValidateErr)
		}
		if len(query) != 0 {
			return x.Model(o).Where(query[0], query[1:]...).UpdateAttrs(values)
		} else {
			return x.Model(o).UpdateAttrs(values)
		}
	})
}

// 更新 model 的非 zero value 字段值到 db
func UpdateModel(tx *db.Session, o Modeler, query ...interface{}) (int64, error) {
	return withTx(tx, func(x *db.Session) (int64, error) {
		if err := o.Validate(); err != nil {
			return 0, e.AutoNew(err, e.DBAttrValidateErr)
		}
		if len(query) == 0 {
			return x.Model(o).Update(o)
		} else if len(query) == 1 {
			return x.Model(o).Where(query[0]).Update(o)
		} else {
			return x.Model(o).Where(query[0], query[1:]...).Update(o)
		}
	})
}

// 更新 model 的所有字段值到 db，即使其值为 zero value(除了 created_at)
// 注意，该方法不会自动更新 updated_at
func UpdateModelAll(tx *db.Session, o Modeler, query ...interface{}) (int64, error) {
	return UpdateModel(tx.Select("*").Omit("created_at"), o, query...)
}

func MustMarshalValue(v interface{}) driver.Value {
	dv, err := MarshalValue(v)
	if err != nil {
		panic(err)
	}
	return dv
}

func MarshalValue(v interface{}) (driver.Value, error) {
	if v == nil {
		return nil, nil
	}
	bs, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return string(bs), nil
}

func UnmarshalValue(src interface{}, dst interface{}) error {
	if src == nil {
		return nil
	}
	var bs []byte
	switch src.(type) {
	case *dmr.DmClob:
		var c dmSchema.Clob
		err := c.Scan(src)
		if err != nil {
			return err
		}
		bs = append(bs[0:0], []byte(c)...)
	case []byte:
		c := src.([]byte)
		bs = append(bs[0:0], c...)
	default:
		return fmt.Errorf("invalid type %T, value: %v", dst, src)
	}
	return json.Unmarshal(bs, dst)
}

func dbMigrate(sess *db.Session) {
	if !autoMigration {
		return
	}

	if err := sess.DropTable("iac_casbin_rule"); err != nil {
		panic(err)
	}
	if err := sess.DropTable("iac__casbin_rule"); err != nil {
		panic(err)
	}
}

var autoMigration = false

func autoMigrate(m Modeler, sess *db.Session) {
	if !autoMigration {
		return
	}

	sess = sess.Model(m)
	if err := sess.GormDB().AutoMigrate(m); err != nil {
		//panic(fmt.Errorf("auto migrate %T: %v", m, err))
		fmt.Errorf("auto migrate %T", m)
		panic(err)
	}

	// 强制修改 table 的字符集和 collate
	if configs.Get().GetDbType() == "mysql" {
		if _, err := sess.Exec(fmt.Sprintf("ALTER TABLE `%s` CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", m.TableName())); err != nil {
			panic(err)
		}
	}

	if err := m.Migrate(sess); err != nil {
		//panic(fmt.Errorf("auto migrate %T: %v", m, err))
		fmt.Errorf("auto migrate %T", m)
		panic(err)
	}

}

func Init(migrate bool) {
	autoMigration = migrate

	var sess *db.Session
	// mysql才设置
	if configs.Get().GetDbType() == "mysql" {
		sess = db.Get().Set("gorm:table_options", "ENGINE=InnoDB CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci").Begin()
	} else {
		sess = db.Get().Begin()
	}
	defer func() {
		logger := logs.Get().WithField("func", "models.Init")
		if r := recover(); r != nil {
			logger.Infof("recover: %v", r)
			if err := sess.Rollback(); err != nil {
				logger.Errorf("rollback error: %v", err)
			}
			panic(r)
		} else if err := sess.Commit(); err != nil {
			logger.Errorf("commit error: %v", err)
			panic(err)
		}
	}()

	autoMigrate(&Organization{}, sess)
	autoMigrate(&Project{}, sess)
	autoMigrate(&Vcs{}, sess)
	autoMigrate(&VcsPr{}, sess)
	autoMigrate(&Template{}, sess)
	autoMigrate(&Env{}, sess)
	autoMigrate(&Resource{}, sess)
	autoMigrate(&ResourceMapping{}, sess)

	autoMigrate(&Variable{}, sess)

	autoMigrate(&Task{}, sess)
	autoMigrate(&ScanTask{}, sess)
	autoMigrate(&TaskStep{}, sess)
	autoMigrate(&DBStorage{}, sess)

	autoMigrate(&User{}, sess)
	autoMigrate(&UserOrg{}, sess)
	autoMigrate(&UserProject{}, sess)

	autoMigrate(&Notification{}, sess)
	autoMigrate(&NotificationEvent{}, sess)
	autoMigrate(&SystemCfg{}, sess)
	autoMigrate(&ResourceAccount{}, sess)
	autoMigrate(&CtResourceMap{}, sess)
	autoMigrate(&OperationLog{}, sess)
	autoMigrate(&Token{}, sess)
	autoMigrate(&Key{}, sess)
	autoMigrate(&TaskComment{}, sess)
	autoMigrate(&ProjectTemplate{}, sess)
	autoMigrate(&Policy{}, sess)
	autoMigrate(&PolicyGroup{}, sess)
	autoMigrate(&PolicyRel{}, sess)
	autoMigrate(&PolicyResult{}, sess)
	autoMigrate(&PolicySuppress{}, sess)
	autoMigrate(&VariableGroup{}, sess)
	autoMigrate(&VariableGroupRel{}, sess)
	autoMigrate(&ResourceDrift{}, sess)
	autoMigrate(&TaskDrift{}, sess)
	autoMigrate(&VariableGroupProjectRel{}, sess)

	autoMigrate(&Bill{}, sess)
	autoMigrate(&BillData{}, sess)
	autoMigrate(&LdapOUOrg{}, sess)
	autoMigrate(&LdapOUProject{}, sess)

	autoMigrate(&UserOperationLog{}, sess)

	autoMigrate(&TagKey{}, sess)
	autoMigrate(&TagValue{}, sess)
	autoMigrate(&TagRel{}, sess)

	dbMigrate(sess)
}
