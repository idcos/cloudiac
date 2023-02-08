// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package e

import (
	"errors"
	"testing"
)

func TestGetAcceptLanguage(t *testing.T) {
	type args struct {
		acceptLanguate string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"cn browser", args{"zh,zh-CN;q=0.9,en-US;q=0.8,en;q=0.7,ko;q=0.6"}, "zh-CN"},
		{"en browser", args{"en-US,en;q=0.9,zh;q=0.8,zh-CN;q=0.7,ko;q=0.6"}, "en-US"},
		{"en", args{"en"}, "en-US"},
		{"en-US", args{"en-US"}, "en-US"},
		{"en-AU", args{"en-AU"}, "en-US"},
		{"default cn", args{""}, "zh-CN"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetAcceptLanguage(tt.args.acceptLanguate); got != tt.want {
				t.Errorf("getAcceptLanguage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestErrorMsg(t *testing.T) {
	type args struct {
		err   Error
		langs string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"cn internal error", args{newError(10000, errors.New("test"), 500), "zh-CN"}, "未知错误"},
		{"en internal error", args{newError(10000, errors.New("test"), 500), "en-US"}, "internal error"},
		{"default cn internal error", args{newError(10000, errors.New("test"), 500), ""}, "未知错误"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ErrorMsg(tt.args.err, tt.args.langs); got != tt.want {
				t.Errorf("ErrorMsg() = %v, want %v", got, tt.want)
			}
		})
	}
}
