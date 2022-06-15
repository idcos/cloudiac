// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTimeMarshalJSON(t *testing.T) {
	/*ANSIC       = "Mon Jan _2 15:04:05 2006"
	UnixDate    = "Mon Jan _2 15:04:05 MST 2006"
	RubyDate    = "Mon Jan 02 15:04:05 -0700 2006"
	RFC822      = "02 Jan 06 15:04 MST"
	RFC822Z     = "02 Jan 06 15:04 -0700" // RFC822 with numeric zone
	RFC850      = "Monday, 02-Jan-06 15:04:05 MST"
	RFC1123     = "Mon, 02 Jan 2006 15:04:05 MST"
	RFC1123Z    = "Mon, 02 Jan 2006 15:04:05 -0700" // RFC1123 with numeric zone
	RFC3339     = "2006-01-02T15:04:05Z07:00"
	RFC3339Nano = "2006-01-02T15:04:05.999999999Z07:00"
	Kitchen     = "3:04PM"
	// Handy time stamps.
	Stamp      = "Jan _2 15:04:05"
	StampMilli = "Jan _2 15:04:05.000"
	StampMicro = "Jan _2 15:04:05.000000"
	StampNano  = "Jan _2 15:04:05.000000000"*/
	tAt1 := time.Now().Format(time.ANSIC)
	tAt2 := time.Now().Format(time.UnixDate)
	tAt3 := time.Now().Format(time.RubyDate)
	tAt4 := time.Now().Format(time.RFC822)
	tAt5 := time.Now().Format(time.RFC822Z)
	tAt6 := time.Now().Format(time.RFC850)
	tAt7 := time.Now().Format(time.RFC1123)
	tAt8 := time.Now().Format(time.RFC1123Z)
	tAt9 := time.Now().Format(time.RFC3339)
	tAt10 := time.Now().Format(time.RFC3339Nano)
	tAt11 := time.Now().Format(time.Stamp)
	tAt12 := time.Now().Format(time.StampMicro)
	tAt13 := time.Now().Format(time.StampMilli)
	tAt14 := time.Now().Format(time.StampNano)
	//t.Logf(tAt)
	timeStringCases := []struct {
		timeAt string
		expect bool
	}{
		{tAt1, false},
		{tAt2, false},
		{tAt3, false},
		{tAt4, false},
		{tAt5, false},
		{tAt6, false},
		{tAt7, false},
		{tAt8, false},
		{tAt9, true},
		{tAt10, true},
		{tAt11, false},
		{tAt12, false},
		{tAt13, false},
		{tAt14, false},
	}

	for _, c := range timeStringCases {
		_, err := Time{}.Parse(c.timeAt)
		if c.expect {
			if err != nil {
				t.Errorf("timdAt %s, err %v", c.timeAt, err)
			}
		} else {
			if err == nil {
				t.Errorf("timdAt %s, err %v", c.timeAt, err)
			}
		}
	}

	tm := Time(time.Now())
	bs, _ := json.Marshal(tm)
	t.Logf("%s", bs)
	tt := Time{}
	if err := json.Unmarshal(bs, &tt); err != nil {
		t.Fatal(err)
	}
	t.Log(tt)
}
