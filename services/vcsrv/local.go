package vcsrv

/*
本地文件系统 vcs 实现
*/

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"cloudiac/utils/logs"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/pkg/errors"
)

type LocalVcs struct {
	basePath string
}

func newLocalVcs(basePath string) *LocalVcs {
	absPath, err := filepath.Abs(basePath)
	if err != nil {
		panic(err)
	}
	return &LocalVcs{basePath: absPath}
}

func (l *LocalVcs) GetRepo(path string) (RepoIface, error) {
	return newLocalRepo(filepath.Join(l.basePath, path))
}

func (l *LocalVcs) ListRepos(namespace string, search string, limit int) ([]RepoIface, error) {
	//logger := logs.Get().WithField("namespace", namespace)

	searchDir := filepath.Join(l.basePath, namespace)
	repoPaths := make([]string, 0)
	err := filepath.WalkDir(searchDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if len(repoPaths) >= limit {
			return filepath.SkipDir
		}

		if !d.IsDir() || p == searchDir {
			return nil
		}

		if matchGlob(search, d.Name()) {
			repoPaths = append(repoPaths, p)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	repos := make([]RepoIface, 0)
	for _, p := range repoPaths {
		if r, err := newLocalRepo(p); err != nil {
			logs.Get().Warnf("open repo '%s' error: %v", p, err)
			continue
		} else {
			repos = append(repos, r)
		}
	}
	return repos, nil
}

type LocalRepo struct {
	path string
	repo *git.Repository
}

func newLocalRepo(path string) (*LocalRepo, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("plain open %s", path))
	}

	return &LocalRepo{
		path: path,
		repo: repo,
	}, nil
}

func (l *LocalRepo) ListBranches(search string, limit int) ([]string, error) {
	refs, err := l.repo.Branches()
	if err != nil {
		return nil, err
	}
	defer refs.Close()

	branches := make([]string, 0)
	err = refs.ForEach(func(ref *plumbing.Reference) error {
		if len(branches) >= limit {
			return nil
		}

		name := filepath.Base(ref.Name().String())
		if matchGlob(search, name) {
			branches = append(branches, name)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return branches, nil
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

func (l *LocalRepo) ListFiles(revision string, searchPath string, search string, recursive bool, limit int) ([]string, error) {
	commit, err := l.getCommit(revision)
	if err != nil {
		return nil, err
	}

	filesIter, err := commit.Files()
	if err != nil {
		return nil, err
	}
	defer filesIter.Close()

	results := make([]string, 0)
	err = filesIter.ForEach(func(file *object.File) error {
		if !strings.HasPrefix(file.Name, searchPath) || len(results) >= limit {
			return nil
		}

		// 非递归时只遍历第一层目录
		if !recursive && (searchPath != "" && filepath.Dir(file.Name) != searchPath) {
			return nil
		}

		if matchGlob(search, filepath.Base(file.Name)) {
			results = append(results, file.Name)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return results, nil
}

func (l *LocalRepo) ReadFileContent(path string) (content []byte, err error) {
	commit, err := l.getCommit("master")
	if err != nil {
		return nil, err
	}

	file, err := commit.File(path)
	if err != nil {
		return nil, err
	}

	c, err := file.Contents()
	return []byte(c), err
}
