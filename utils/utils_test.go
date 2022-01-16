// Copyright 2021 CloudJ Company Limited. All rights reserved.

package utils

import (
	"net/url"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGlobMatch(t *testing.T) {
	cases := []struct {
		pattern string
		str     string
		expect  bool
	}{
		{"/a/b/c/d", "/a/b/c/d", true},
		{"/a/b/c/d", "/a/b/c/", false},
		{"/a/b/c", "/a/b/c/d", false},
		{"/a/b/c/*", "/a/b/c/d", true},
		{"/a/b/c/*", "/a/b/c/", true},
		{"/a/b/c/*", "/a/b/c/dd", true},
		{"/a/b/c/?", "/a/b/c/d", true},
		{"/a/b/c/?", "/a/b/c/dd", false},
		{"/a/b/*/d", "/a/b/cc/d", true},
		{"/a/b/?/d", "/a/b/c/d", true},
		{"/a/b/*/z", "/a/b/c/d/e/f/z", false},
		// 目前使用的 filepath.Match() 该函数的匹配规则不支持使用 * 来指代多层目录
		{"/a/b/**/z", "/a/b/c/d/e/f/z", false},
	}

	for _, c := range cases {
		match, err := GlobMatch(c.pattern, c.str)
		if err != nil {
			t.Fatal(err)
		}
		if match != c.expect {
			t.Fatalf("%s, %s, expect %v, got %v", c.pattern, c.str, c.expect, match)
		}
	}
}

func TestFileExists(t *testing.T) {
	cases := []struct {
		path   string
		expect bool
	}{
		{"/not/this/filepath", false},
		{os.TempDir(), true},
	}

	for _, c := range cases {
		v := FileExist(c.path)
		assert.Equal(t, c.expect, v)
	}
}

func TestAesEncrypt(t *testing.T) {
	text := "xxx"
	key := "W5ds1zjYGHhh71dCOMMy5bG5ellAzQxx"
	ss, err := AesEncryptWithKey(text, key)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ss)
	ds, err := AesDecryptWithKey(ss, key)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, text, ds)
}

func TestGetUrlParams(t *testing.T) {
	type args struct {
		uri string
	}
	tests := []struct {
		name string
		args args
		want url.Values
	}{
		{name: "invalidUrl", args: args{uri: "test error uri"}, want: url.Values{}},
		{name: "validUrl", args: args{uri: "http://10.0.0.1?key=xxxxxx"}, want: url.Values{"key": []string{"xxxxxx"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetUrlParams(tt.args.uri); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetUrlParams() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJoinUrl(t *testing.T) {
	cases := []struct {
		elems  []string
		except string
	}{
		{[]string{"http://example.com", "a"}, "http://example.com/a"},
		{[]string{"http://example.com", "/a"}, "http://example.com/a"},
		{[]string{"http://example.com/", "a"}, "http://example.com/a"},
		{[]string{"http://example.com/", "a", "b"}, "http://example.com/a/b"},
		{[]string{"http://example.com/", "a", "/b"}, "http://example.com/a/b"},
		{[]string{"http://example.com/", "/a", "b"}, "http://example.com/a/b"},
		{[]string{"http://example.com/", "/a", "/b"}, "http://example.com/a/b"},
		{[]string{"http://example.com/", "", "/b"}, "http://example.com/b"},
		{[]string{"http://example.com/", "", "b"}, "http://example.com/b"},
	}
	for _, c := range cases {
		url := JoinURL(c.elems[0], c.elems[1:]...)
		if url != c.except {
			t.Fatalf("join url: %v, got: '%s', except: '%s'", c.elems, url, c.except)
		}
	}
}
