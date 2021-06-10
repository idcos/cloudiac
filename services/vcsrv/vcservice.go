package vcsrv

import (
	"cloudiac/consts"
	"cloudiac/consts/e"
	"cloudiac/models"
	"path"

	"github.com/pkg/errors"
)

/*
version control service 接口
*/

type VcsIfaceOptions struct {
	Ref       string
	Path      string
	Search    string
	Recursive bool
	Limit     int
	Offset    int
}

type VcsIface interface {
	// GetRepo 列出仓库
	// param idOrPath: 仓库id或者路径
	GetRepo(idOrPath string) (RepoIface, error)

	// ListRepos 列出仓库列表
	// param namespace: namespace 可用于表示用户、组织等
	// param search: 搜索字符串
	// param limit: 限制返回的文件数，传 0 表示无限制
	ListRepos(namespace, search string, limit, offset uint) ([]RepoIface, int64, error)
}

type RepoIface interface {
	// ListBranches
	// param search: 搜索字符串
	// param limit: 限制返回的文件数，传 0 表示无限制
	// param offset: 偏移量
	ListBranches(search string, limit, offset uint) ([]string, error)

	// BranchCommitId
	//param branch: 分支
	BranchCommitId(branch string) (string, error)

	// ListFiles 列出指定路径下的文件
	// param ref: 分支或者 commit id
	// param filename: 文件名称部分的点
	// param path: 路径
	// param search: 搜索字符串
	// param recursive: 是否递归遍历子目录
	// param limit: 限制返回的文件数，传 0 表示无限制
	// return: 返回文件路径列表，路径为完整路径(即包含传入的 path 部分)
	ListFiles(option VcsIfaceOptions) ([]string, error)

	// ReadFileContent
	// param path: 路径
	// param branch: 分支
	ReadFileContent(branch, path string) (content []byte, err error)

	// FormatRepoSearch 格式化输出前端需要的内容
	FormatRepoSearch() (project *Projects, err e.Error)
}

func GetVcsInstance(vcs *models.Vcs) (VcsIface, error) {
	switch vcs.VcsType {
	case consts.GitTypeLocal:
		return newLocalVcs(vcs.Address), nil
	case consts.GitTypeGitLab:
		return newGitlabInstance(vcs)
	case consts.GitTypeGitEA:
		return newGiteaInstance(vcs)
	//case consts.GitTypeGithub:
	//	//return newGitlabInstance(vcs)
	//case consts.GitTypeGitee:
	//	//return newGitlabInstance(vcs)
	default:
		return nil, errors.New("vcs type doesn't exist")
	}
}

func matchGlob(pattern, name string) bool {
	if pattern == "" {
		return true
	}

	matched, err := path.Match(pattern, name)
	if err != nil {
		return false
	}
	return matched
}
