package models

import (
	"cloudiac/configs"
	"database/sql/driver"
	"fmt"
	"github.com/jiangliuhong/gorm-driver-dm/dmr"
)

type ByteBlob []byte

func (b ByteBlob) Value() (driver.Value, error) {
	if len(b) == 0 {
		return nil, nil
	}
	dbType := configs.Get().GetDbType()
	if dbType == "dameng" {
		return dmr.NewBlob(b), nil
	} else {
		return []byte(b), nil
	}
}

func (b *ByteBlob) Scan(value interface{}) error {
	if value == nil {
		*b = nil
		return nil
	}

	switch vt := value.(type) {
	case *dmr.DmBlob:
		dmBlob := value.(*dmr.DmBlob)
		length, err := dmBlob.GetLength()
		if err != nil {
			return err
		}
		rb := make([]byte, length)
		_, err = dmBlob.Read(rb)
		if err != nil {
			return err
		}
		*b = append((*b)[0:0], rb...)
		return nil
	case []byte:
		bs := value.([]byte)
		*b = append((*b)[0:0], bs...)
		return nil
	default:
		return fmt.Errorf("invalid type %T, value: %v", vt, value)
	}
}
