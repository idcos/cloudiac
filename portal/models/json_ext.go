// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package models

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"github.com/jiangliuhong/gorm-driver-dm/dmr"
	dmSchema "github.com/jiangliuhong/gorm-driver-dm/schema"
)

type JSON []byte

func (j JSON) Value() (driver.Value, error) {
	if j.IsNull() {
		return nil, nil
	}
	return string(j), nil
}

func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	switch vt := value.(type) {
	case *dmr.DmClob:
		var c dmSchema.Clob
		err := c.Scan(value)
		if err != nil {
			return err
		}
		*j = append((*j)[0:0], []byte(c)...)
		return nil
	case []byte:
		bs := value.([]byte)
		*j = append((*j)[0:0], bs...)
		return nil
	default:
		return fmt.Errorf("invalid type %T, value: %v", vt, value)
	}
}

func (m JSON) MarshalJSON() ([]byte, error) {
	if m == nil {
		return []byte("null"), nil
	}
	return m, nil
}

func (m *JSON) UnmarshalJSON(data []byte) error {
	if m == nil {
		return errors.New("null point exception")
	}
	*m = append((*m)[0:0], data...)
	return nil
}

func (j JSON) IsNull() bool {
	return len(j) == 0 || string(j) == "null"
}
