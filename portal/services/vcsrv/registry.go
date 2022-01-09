package vcsrv

/*
registry vcs 实现
*/

import (
	"bytes"
	"cloudiac/portal/consts/e"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
)

type RegistryVcs struct {
	RegitryAddr string
}

func newRegistryVcs(registryAddr string) *RegistryVcs {
	return &RegistryVcs{RegitryAddr: registryAddr}
}

// TODO
func (rv *RegistryVcs) GetRepo(path string) (RepoIface, error) {
	return &RegistryRepo{VcsAddr: rv.RegitryAddr, RepoPath: path}, nil
}

// TODO
func (rv *RegistryVcs) ListRepos(namespace string, search string, limit, offset int) ([]RepoIface, int64, error) {

	return nil, 0, nil
}

type RegistryRepo struct {
	VcsAddr  string
	RepoPath string // vcs 下的相对路径
}

func (r *RegistryRepo) ListBranches() ([]string, error) {
	path := fmt.Sprintf("%s/api/v1/vcs/repo/branches", r.VcsAddr)
	_, body, err := registryVcsRequest(path, "GET", map[string]string{
		"path": r.RepoPath,
	})
	if err != nil {
		return nil, err
	}

	var resp struct {
		Result struct {
			Branches []string `json:"branches"`
		} `json:"result"`
	}

	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, err
	}

	return resp.Result.Branches, nil
}

func (r *RegistryRepo) ListTags() ([]string, error) {
	path := fmt.Sprintf("%s/api/v1/vcs/repo/tags", r.VcsAddr)
	_, body, err := registryVcsRequest(path, "GET", map[string]string{
		"path": r.RepoPath,
	})
	if err != nil {
		return nil, err
	}

	var resp struct {
		Result struct {
			Tags []string `json:"tags"`
		} `json:"result"`
	}

	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, err
	}

	return resp.Result.Tags, nil
}

func (r *RegistryRepo) BranchCommitId(branch string) (string, error) {
	path := fmt.Sprintf("%s/api/v1/vcs/repo/branch_commit_id", r.VcsAddr)
	_, body, err := registryVcsRequest(path, "GET", map[string]string{
		"path":   r.RepoPath,
		"branch": branch,
	})
	if err != nil {
		return "", err
	}

	var resp struct {
		Result struct {
			CommitId string `json:"commitId"`
		} `json:"result"`
	}

	err = json.Unmarshal(body, &resp)
	if err != nil {
		return "", err
	}

	return resp.Result.CommitId, nil
}

func (r *RegistryRepo) ListFiles(opt VcsIfaceOptions) ([]string, error) {
	path := fmt.Sprintf("%s/api/v1/vcs/repo/files", r.VcsAddr)
	recursive := "false"
	if opt.Recursive {
		recursive = "true"
	}
	_, body, err := registryVcsRequest(path, "GET", map[string]string{
		"path":      r.RepoPath,
		"ref":       opt.Ref,
		"filePath":  opt.Path,
		"search":    opt.Search,
		"recursive": recursive,
		"limit":     fmt.Sprintf("%d", opt.Limit),
		"offset":    fmt.Sprintf("%d", opt.Offset),
	})
	if err != nil {
		return nil, err
	}

	var resp struct {
		Result struct {
			Files []string `json:"files"`
		} `json:"result"`
	}

	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, err
	}

	return resp.Result.Files, nil
}

func (r *RegistryRepo) ReadFileContent(revision string, filePath string) (content []byte, err error) {
	path := fmt.Sprintf("%s/api/v1/vcs/repo/file_content", r.VcsAddr)
	_, body, err := registryVcsRequest(path, "GET", map[string]string{
		"path":     r.RepoPath,
		"branch":   revision,
		"filePath": filePath,
	})
	if err != nil {
		return nil, err
	}

	var resp struct {
		Result struct {
			Content string `json:"content"`
		} `json:"result"`
	}

	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, err
	}

	return []byte(resp.Result.Content), nil
}

func (r *RegistryRepo) FormatRepoSearch() (*Projects, e.Error) {
	path := fmt.Sprintf("%s/api/v1/vcs/repo/info", r.VcsAddr)
	_, body, err := registryVcsRequest(path, "GET", map[string]string{
		"path": r.RepoPath,
	})
	if err != nil {
		return nil, e.New(e.InternalError, err)
	}

	var resp struct {
		Result *Projects `json:"result"`
	}

	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, e.New(e.InternalError, err)
	}

	return resp.Result, nil
}

func (r *RegistryRepo) DefaultBranch() string {
	path := fmt.Sprintf("%s/api/v1/vcs/repo/default_branch", r.VcsAddr)
	_, body, err := registryVcsRequest(path, "GET", map[string]string{
		"path": r.RepoPath,
	})
	if err != nil {
		return ""
	}

	var resp struct {
		Result struct {
			DefaultBranch string `json:"defaultBranch"`
		} `json:"result"`
	}

	err = json.Unmarshal(body, &resp)
	if err != nil {
		return ""
	}

	return resp.Result.DefaultBranch
}

func (r *RegistryRepo) AddWebhook(url string) error {
	return nil
}

func (r *RegistryRepo) ListWebhook() ([]ProjectsHook, error) {
	ph := make([]ProjectsHook, 0)
	return ph, nil
}

func (r *RegistryRepo) DeleteWebhook(id int) error {
	return nil
}

func (r *RegistryRepo) CreatePrComment(prId int, comment string) error {

	return nil
}

func registryVcsRequest(path, method string, params map[string]string) (*http.Response, []byte, error) {
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	for k, v := range params {
		_ = writer.WriteField(k, v)
	}
	err := writer.Close()
	if err != nil {
		return nil, nil, err
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, path, payload)

	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	return res, body, err
}
