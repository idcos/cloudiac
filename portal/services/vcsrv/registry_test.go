// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package vcsrv

import (
	"cloudiac/portal/models"
	"fmt"
	"testing"
)

// 该测试需要 registry侧服务的配合

func getTestRepo() RepoIface {
	var addr = "http://localhost:9233"
	var testRepo = "test/myrepo"

	vcs, err := newRegistryVcs(&models.Vcs{
		Address: addr,
	})
	if err != nil {
		panic(err)
	}
	repo, _ := vcs.GetRepo(testRepo)
	return repo
}

func TestListBranches(t *testing.T) {
	repo := getTestRepo()

	branches, err := repo.ListBranches()
	if err != nil {
		t.Error(err)
	}
	fmt.Println(branches)
}

func TestListTags(t *testing.T) {
	repo := getTestRepo()

	branches, err := repo.ListTags()
	if err != nil {
		t.Error(err)
	}
	fmt.Println(branches)
}

func TestBranchCommitId(t *testing.T) {
	repo := getTestRepo()

	commitId, err := repo.BranchCommitId("master")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(commitId)
}

func TestListFiles(t *testing.T) {
	repo := getTestRepo()

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
	repo := getTestRepo()

	content, err := repo.ReadFileContent("master", "aaa")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(string(content))
}
func TestFormatRepoSearch(t *testing.T) {
	repo := getTestRepo()

	proj, err := repo.FormatRepoSearch()
	if err != nil {
		t.Error(err)
	}
	fmt.Println(proj)
}
func TestDefaultBranch(t *testing.T) {
	repo := getTestRepo()

	branch := repo.DefaultBranch()
	fmt.Println(branch)
}
