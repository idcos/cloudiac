package vcsrv

import "path"

/*
version control service 接口
*/

type VcsIface interface {
	GetRepo(idOrPath string) (RepoIface, error)
	// ListRepos 列出仓库列表
	// param namespace: namespace 可用于表示用户、组织等
	// param search: 搜索字符串
	// param limit: 限制返回的文件数，传 0 表示无限制
	ListRepos(namespace string, search string, limit int) ([]RepoIface, error)
}

type RepoIface interface {
	ListBranches(search string, limit int) ([]string, error)
	BranchCommitId(branch string) (string, error)

	// ListFiles 列出指定路径下的文件
	// param revision: git revision (分支或者 commit id)
	// param path: 路径
	// param search: 搜索字符串
	// param recursive: 是否递归遍历子目录
	// param limit: 限制返回的文件数，传 0 表示无限制
	// return: 返回文件路径列表，路径为完整路径(即包含传入的 path 部分)
	ListFiles(revision string, path string, search string, recursive bool, limit int) ([]string, error)

	ReadFileContent(path string) (content []byte, err error)
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
