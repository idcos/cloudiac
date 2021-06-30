package vcsrv

import (
	"path/filepath"
	"strings"
	"testing"

	"cloudiac/portal/consts"

	"github.com/stretchr/testify/assert"
)

var (
	testLocalVcs = newLocalVcs(filepath.Join("../../", consts.LocalGitReposPath))
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

func TestLocalVcs(t *testing.T) {
	assert := assert.New(t)
	repos, _, err := testLocalVcs.ListRepos("cloud-iac", "", 1, 0)
	assert.NoError(err)
	if !assert.Equal(1, len(repos)) {
		t.FailNow()
	}

	repo := repos[0]

	basePath := strings.Replace(repo.(*LocalRepo).path, testLocalVcs.absPath, "", -1)
	repo, err = testLocalVcs.GetRepo(basePath)
	assert.NoError(err)

	branches, err := repo.ListBranches()
	assert.NoError(err)
	if !assert.Equal(1, len(branches)) {
		t.Failed()
	}

	for _, b := range branches {
		commitId, err := repo.BranchCommitId(b)
		assert.NoError(err)
		t.Logf("branch %v, commit %v", b, commitId)

		{
			// 测试 limit 和 offset
			files, err := repo.ListFiles(VcsIfaceOptions{
				Ref:       commitId,
				Path:      "",
				Search:    "",
				Recursive: true,
				Limit:     1,
				Offset:    1,
			})
			assert.NoError(err)
			if !assert.Equal(1, len(files)) {
				t.Failed()
			}
		}

		files, err := repo.ListFiles(VcsIfaceOptions{
			Ref:       commitId,
			Path:      "",
			Search:    "*.tf",
			Recursive: false,
			Limit:     1,
			Offset:    0,
		})
		assert.NoError(err)
		if !assert.Equal(1, len(files)) {
			t.Failed()
		}

		content, err := repo.ReadFileContent(b, files[0])
		assert.NoError(err)
		t.Logf("%s content: %s", files[0], content)
	}
}
