// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTimeMarshalJSON(t *testing.T) {
	ts := "2022-07-03T06:15:55.000Z"
	tt1, err := Time{}.Parse(ts)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(tt1)

	tm := Time(time.Now())
	bs, _ := json.Marshal(tm)
	t.Logf("%s", bs)
	tt2 := Time{}
	if err := json.Unmarshal(bs, &tt1); err != nil {
		t.Fatal(err)
	}
	t.Log(tt2)
}
