package utils

import "testing"

func TestGlobMatch(t *testing.T) {
	cases := []struct {
		pattern string
		str     string
		expect  bool
	}{
		{ "/a/b/c/d", "/a/b/c/d", true},
		{ "/a/b/c/d", "/a/b/c/", false},
		{ "/a/b/c", "/a/b/c/d", false},
		{ "/a/b/c/*", "/a/b/c/d", true},
		{ "/a/b/c/*", "/a/b/c/", true},
		{ "/a/b/c/*", "/a/b/c/dd", true},
		{ "/a/b/c/?", "/a/b/c/d", true},
		{ "/a/b/c/?", "/a/b/c/dd", false},
		{ "/a/b/*/d", "/a/b/cc/d", true},
		{ "/a/b/?/d", "/a/b/c/d", true},
		{ "/a/b/*/z", "/a/b/c/d/e/f/z", false},
		// 目前使用的 filepath.Match() 该函数的匹配规则不支持使用 * 来指代多层目录
		{ "/a/b/**/z", "/a/b/c/d/e/f/z", false},
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
