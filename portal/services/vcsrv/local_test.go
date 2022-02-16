// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package vcsrv

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatchGlob(t *testing.T) {
	cases := []struct {
		name   string
		search string
		except bool
	}{
		{"abc", "", true},
		{"abc", "a", true},
		{"abc", "ab", true},
		{"abc", "abc", true},
		{"abc", "bc", true},
		{"abc", "abcd", false},
		{"abc.def", ".def", true},
		{"ab.cdef", ".def", false},
		// path.Match() 不匹配 "/"
		{"a/b/c", "/a/b/", false},
		{"a/b/c", "/a/b/c", false},
	}

	for _, c := range cases {
		assert.Equal(t, c.except, matchGlob(c.search, c.name), "%v", c)
	}
}
