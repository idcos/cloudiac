package models

import (
	"fmt"
	"github.com/jiangliuhong/gorm-driver-dm/dmr"
	dmSchema "github.com/jiangliuhong/gorm-driver-dm/schema"
)

type Text string

func (t *Text) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		*t = Text(v)
		return nil
	case []byte:
		*t = Text(v)
		return nil
	case *dmr.DmClob:
		var c dmSchema.Clob
		err := c.Scan(value)
		if err != nil {
			return err
		}
		*t = Text(c)
		return nil
	}
	return fmt.Errorf("unsupported Scan, storing driver.Value type %T into type Text", value)
}
