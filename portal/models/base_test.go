// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTimeMarshalJSON(t *testing.T) {
	tm := Time(time.Now())
	bs, _ := json.Marshal(tm)
	t.Logf("%s", bs)
}
