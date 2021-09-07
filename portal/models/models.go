// Copyright 2021 CloudJ Company Limited. All rights reserved.

package models

import (
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

//CreateBatch Fixme 目前切片Modeler类型无法与批量插入公用
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
