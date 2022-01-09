package vcsrv

import (
	"fmt"
	"testing"
)

// 改测试需要 registry侧服务的配合

func TestListBranches(t *testing.T) {
	vcs := newRegistryVcs("http://localhost:9233")
	repo, _ := vcs.GetRepo("test/myrepo")

	branches, err := repo.ListBranches()
	if err != nil {
		t.Error(err)
	}
	fmt.Println(branches)
}

func TestListTags(t *testing.T) {
	vcs := newRegistryVcs("http://localhost:9233")
	repo, _ := vcs.GetRepo("test/myrepo")

	branches, err := repo.ListTags()
	if err != nil {
		t.Error(err)
	}
	fmt.Println(branches)
}

func TestBranchCommitId(t *testing.T) {
	vcs := newRegistryVcs("http://localhost:9233")
	repo, _ := vcs.GetRepo("test/myrepo")

	commitId, err := repo.BranchCommitId("master")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(commitId)
}

func TestListFiles(t *testing.T) {
	vcs := newRegistryVcs("http://localhost:9233")
	repo, _ := vcs.GetRepo("test/myrepo")

	var opt = VcsIfaceOptions{
		Ref: "master",
	}

	fmt.Println(opt)
	files, err := repo.ListFiles(opt)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(files)
}
func TestFileContent(t *testing.T) {
	vcs := newRegistryVcs("http://localhost:9233")
	repo, _ := vcs.GetRepo("test/myrepo")

	content, err := repo.ReadFileContent("master", "aaa")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(string(content))
}
func TestFormatRepoSearch(t *testing.T) {
	vcs := newRegistryVcs("http://localhost:9233")
	repo, _ := vcs.GetRepo("test/myrepo")

	proj, err := repo.FormatRepoSearch()
	if err != nil {
		t.Error(err)
	}
	fmt.Println(proj)
}
func TestDefaultBranch(t *testing.T) {
	vcs := newRegistryVcs("http://localhost:9233")
	repo, _ := vcs.GetRepo("test/myrepo")

	branch := repo.DefaultBranch()
	fmt.Println(branch)
}
