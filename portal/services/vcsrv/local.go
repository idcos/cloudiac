// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package vcsrv

/*
本地文件系统 vcs 实现
*/

import (
	"cloudiac/configs"
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/utils/logs"
	"fmt"
	"io/fs"
	"path"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/storer"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/pkg/errors"
)

type LocalVcs struct {
	absPath string // 文件系统绝对路径
}

func newLocalVcs(basePath string) *LocalVcs {
	absPath, err := filepath.Abs(basePath)
	if err != nil {
		panic(err)
	}
	return &LocalVcs{absPath: absPath}
}

func (l *LocalVcs) GetRepo(path string) (RepoIface, error) {
	return newLocalRepo(l.absPath, path)
}

func (l *LocalVcs) ListRepos(namespace string, search string, limit, offset int) ([]RepoIface, int64, error) {
	//logger := logs.Get().WithField("namespace", namespace)

	searchDir := filepath.Join(l.absPath, namespace)
	repoPaths := make([]string, 0)
	err := filepath.WalkDir(searchDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if p == searchDir || !d.IsDir() || !strings.HasSuffix(d.Name(), ".git") {
			return nil
		}

		if matchGlob(search, d.Name()) {
			repoPaths = append(repoPaths, strings.TrimPrefix(p, l.absPath))
		}
		return nil
	})
	if err != nil {
		return nil, 0, err
	}
	var total = int64(len(repoPaths))
	repoPaths = repoPaths[offset:]
	if limit != 0 && len(repoPaths) > limit {
		repoPaths = repoPaths[:limit]
	}

	repos := make([]RepoIface, 0)
	for _, p := range repoPaths {
		if r, err := newLocalRepo(l.absPath, p); err != nil {
			logs.Get().Warnf("open repo '%s' error: %v", p, err)
			continue
		} else {
			repos = append(repos, r)
		}
	}
	return repos, total, nil
}

func (l *LocalVcs) UserInfo() (UserInfo, error) {

	return UserInfo{}, nil
}

func (l *LocalVcs) TokenCheck() error {
	return nil
}

type LocalRepo struct {
	absPath string // 文件系统中的绝对路径
	path    string // vcs 下的相对路径
	repo    *git.Repository
}

func newLocalRepo(dir string, path string) (*LocalRepo, error) {
	absPath := filepath.Join(dir, path)
	repo, err := git.PlainOpen(absPath)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("plain open %s", path))
	}

	return &LocalRepo{
		absPath: absPath,
		path:    path,
		repo:    repo,
	}, nil
}

func getRevision(refs storer.ReferenceIter) ([]string, error) {
	defer refs.Close()

	branches := make([]string, 0)
	err := refs.ForEach(func(ref *plumbing.Reference) error {
		branches = append(branches, ref.Name().Short())
		return nil
	})
	if err != nil {
		return nil, err
	}
	return branches, nil
}

func (l *LocalRepo) ListBranches() ([]string, error) {
	refs, err := l.repo.Branches()
	if err != nil {
		return nil, err
	}
	return getRevision(refs)

}

func (l *LocalRepo) ListTags() ([]string, error) {
	refs, err := l.repo.Tags()
	if err != nil {
		return nil, err
	}
	return getRevision(refs)

}

func (l *LocalRepo) BranchCommitId(branch string) (string, error) {
	hash, err := l.repo.ResolveRevision(plumbing.Revision(branch))
	if err != nil {
		return "", err
	}
	return hash.String(), nil
}

func (l *LocalRepo) getCommit(revision string) (*object.Commit, error) {
	if revision == "" {
		return nil, fmt.Errorf("invalid revision '%v'", revision)
	}
	hash, err := l.repo.ResolveRevision(plumbing.Revision(revision))
	if err != nil {
		return nil, err
	}

	return l.repo.CommitObject(*hash)
}

func getMatchedFiles(filesIter *object.FileIter, opt VcsIfaceOptions) ([]string, error) {
	results := make([]string, 0)
	err := filesIter.ForEach(func(file *object.File) error {
		if !strings.HasPrefix(file.Name, opt.Path) {
			return nil
		}

		if !opt.Recursive {
			// 非递归时只遍历第一层目录
			if (opt.Path == "" && filepath.Dir(file.Name) != ".") ||
				(opt.Path != "" && filepath.Dir(file.Name) != opt.Path) {
				return nil
			}
		}

		if matchGlob(opt.Search, filepath.Base(file.Name)) {
			results = append(results, file.Name)
		}
		return nil
	})

	return results, err
}

func (l *LocalRepo) ListFiles(opt VcsIfaceOptions) ([]string, error) {
	branch := getBranch(l, opt.Ref)
	commit, err := l.getCommit(branch)
	if err != nil {
		return nil, err
	}

	filesIter, err := commit.Files()
	if err != nil {
		return nil, err
	}
	defer filesIter.Close()

	results, err := getMatchedFiles(filesIter, opt)
	if err != nil {
		return nil, err
	}

	results = results[opt.Offset:]
	if opt.Limit != 0 {
		results = results[:opt.Limit]
	}
	return results, nil
}

func (l *LocalRepo) ReadFileContent(revision string, path string) (content []byte, err error) {
	commit, err := l.getCommit(revision)
	if err != nil {
		return nil, err
	}

	file, err := commit.File(path)
	if err != nil {
		if strings.Contains(err.Error(), "file not found") {
			return nil, e.New(e.ObjectNotExists)
		}
		return nil, err
	}

	c, err := file.Contents()
	return []byte(c), err
}

func (l *LocalRepo) FormatRepoSearch() (*Projects, e.Error) {
	head, err := l.repo.Head()
	if err != nil {
		return nil, e.New(e.InternalError, err)
	}

	headCommit, err := l.repo.CommitObject(head.Hash())
	if err != nil {
		return nil, e.New(e.InternalError, err)
	}
	httpUrl := fmt.Sprintf("%s/%s",
		strings.Trim(configs.Get().Portal.Address, "/"),
		strings.Trim(path.Join(consts.ReposUrlPrefix, l.path), "/"))

	return &Projects{
		ID:             l.path,
		Description:    "",
		DefaultBranch:  l.DefaultBranch(),
		SSHURLToRepo:   "",
		HTTPURLToRepo:  httpUrl,
		Name:           strings.TrimSuffix(filepath.Base(l.path), ".git"),
		LastActivityAt: &headCommit.Author.When,
		FullName:       strings.TrimSuffix(filepath.Base(l.path), ".git"),
	}, nil
}

func (l *LocalRepo) DefaultBranch() string {
	head, _ := l.repo.Head()
	return head.Name().Short()
}

func (l *LocalRepo) AddWebhook(url string) error {
	return nil
}

func (l *LocalRepo) ListWebhook() ([]RepoHook, error) {
	return nil, nil
}

func (l *LocalRepo) DeleteWebhook(id int) error {
	return nil
}

func (l *LocalRepo) CreatePrComment(prId int, comment string) error {

	return nil
}

func (l *LocalRepo) GetFullFilePath(address, filePath, repoRevision string) string {
	return ""
}

func (l *LocalRepo) GetCommitFullPath(address, commitId string) string {
	return ""
}
