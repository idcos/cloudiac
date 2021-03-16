package models

import (
	"fmt"

	"cloudiac/consts/e"
	"cloudiac/libs/db"
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
			return x.Model(o).Where(query[0], query[1:]...).Update(values)
		} else {
			return x.Model(o).Update(values)
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

var autoMigration = false

func autoMigrate(m Modeler, sess *db.Session) {
	if !autoMigration {
		return
	}

	sess = sess.Model(m)
	if err := sess.DB().AutoMigrate(m).Error; err != nil {
		panic(fmt.Errorf("auto migrate %T: %v", m, err))
	}
	if err := m.Migrate(sess); err != nil {
		panic(fmt.Errorf("auto migrate %T: %v", m, err))
	}
}

func Init(migrate bool) {
	autoMigration = migrate

	sess := db.ToSess(db.Get().DB().Set("gorm:table_options", "ENGINE=InnoDB")).Begin().Debug()
	defer func() {
		if r := recover(); r != nil {
			_ = sess.Rollback()
			panic(r)
		} else if err := sess.Commit(); err != nil {
			panic(err)
		}
	}()

	autoMigrate(&User{}, sess)
}
