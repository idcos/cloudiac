// Copyright 2021 CloudJ Company Limited. All rights reserved.

package models

import (
	"cloudiac/utils/logs"
	"database/sql/driver"
	"encoding/json"
	"fmt"

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

func getSess(tx *db.Session) *db.Session {
	if tx == nil {
		return db.Get()
	}
	return tx
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

func CreateBatch(tx *db.Session, o []Modeler) error {
	_, err := withTx(tx, func(x *db.Session) (int64, error) {
		for _, v := range o {
			if err := v.Validate(); err != nil {
				return 0, err
			}
		}
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
			return 0, e.New(e.DBAttrValidateErr, err)
		}
		if len(query) != 0 {
			return x.Model(o).Where(query[0], query[1:]...).UpdateAttrs(values)
		} else {
			return x.Model(o).UpdateAttrs(values)
		}
	})
}

func UpdateModel(tx *db.Session, o Modeler, query ...interface{}) (int64, error) {
	return withTx(tx, func(x *db.Session) (int64, error) {
		if err := o.Validate(); err != nil {
			return 0, e.New(e.DBAttrValidateErr, err)
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

	bs, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("invalid type %T, value: %T", src, src)
	}
	return json.Unmarshal(bs, dst)
}

var autoMigration = false

func autoMigrate(m Modeler, sess *db.Session) {
	if !autoMigration {
		return
	}

	sess = sess.Model(m)
	if err := sess.GormDB().AutoMigrate(m); err != nil {
		panic(fmt.Errorf("auto migrate %T: %v", m, err))
	}
	if err := m.Migrate(sess); err != nil {
		panic(fmt.Errorf("auto migrate %T: %v", m, err))
	}
}

func Init(migrate bool) {
	autoMigration = migrate

	sess := db.Get().Set("gorm:table_options", "ENGINE=InnoDB").Begin().Debug()
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
	autoMigrate(&Template{}, sess)
	autoMigrate(&Env{}, sess)
	autoMigrate(&Resource{}, sess)

	autoMigrate(&Variable{}, sess)

	autoMigrate(&Task{}, sess)
	autoMigrate(&TaskStep{}, sess)
	autoMigrate(&DBStorage{}, sess)

	autoMigrate(&User{}, sess)
	autoMigrate(&UserOrg{}, sess)
	autoMigrate(&UserProject{}, sess)

	autoMigrate(&NotificationCfg{}, sess)
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
}
