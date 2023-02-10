// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package vcsrv

import (
	"bytes"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/models"
	"cloudiac/utils"
	"cloudiac/utils/logs"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	git "github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// newGithubInstance
// github api文档: https://docs.github.com/cn/rest/reference/repos
func newGithubInstance(vcs *models.Vcs) (VcsIface, error) {
	return &githubVcs{vcs: vcs}, nil

}

type githubVcs struct {
	vcs *models.Vcs
}

func (github *githubVcs) GetRepo(idOrPath string) (RepoIface, error) {
	path := utils.GenQueryURL(github.vcs.Address, fmt.Sprintf("/repos/%s", idOrPath), nil)
	_, body, er := githubRequest(path, "GET", github.vcs.VcsToken, nil)
	if er != nil {
		return nil, e.New(e.VcsError, er)
	}

	rep := RepositoryGithub{}
	_ = json.Unmarshal(body, &rep)
	return &githubRepoIface{
		vcs:        github.vcs,
		repository: &rep,
	}, nil

}

type RepositoryGithub struct {
	ID            int64     `json:"id"`
	Description   string    `json:"description"`
	DefaultBranch string    `json:"default_branch"`
	SSHURL        string    `json:"ssh_url"`
	CloneURL      string    `json:"clone_url"`
	Name          string    `json:"name"`
	Updated       time.Time `json:"updated_at"`
	FullName      string    `json:"full_name"`
}

func (github *githubVcs) ListRepos(namespace, search string, limit, offset int) ([]RepoIface, int64, error) {
	page := utils.LimitOffset2Page(limit, offset)
	urlParam := url.Values{}
	urlParam.Set("page", strconv.Itoa(page))
	urlParam.Set("per_page", strconv.Itoa(limit))

	if search != "" {
		urlParam.Set("q", search)
	}
	path := utils.GenQueryURL(github.vcs.Address, "/user/repos", urlParam)
	_, body, err := githubRequest(path, "GET", github.vcs.VcsToken, nil)
	if err != nil {
		return nil, 0, e.New(e.VcsError, err)
	}

	token, err := GetVcsToken(github.vcs.VcsToken)
	if err != nil {
		return nil, 0, err
	}
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := git.NewClient(tc)
	_, r, err := client.Repositories.List(ctx, "", &git.RepositoryListOptions{
		ListOptions: git.ListOptions{
			Page:    1,
			PerPage: 1,
		},
	})
	if err != nil {
		return nil, 0, e.New(e.VcsError, err)
	}

	rep := make([]*RepositoryGithub, 0)
	_ = json.Unmarshal(body, &rep)
	repoList := make([]RepoIface, 0)
	for _, v := range rep {
		repoList = append(repoList, &githubRepoIface{
			vcs:        github.vcs,
			repository: v,
			total:      r.LastPage,
		})
	}

	return repoList, int64(r.LastPage), nil
}

func (github *githubVcs) UserInfo() (UserInfo, error) {

	return UserInfo{}, nil
}

func (github *githubVcs) TokenCheck() error {
	limit, offset := 1, 1
	page := utils.LimitOffset2Page(limit, offset)
	urlParam := url.Values{}
	urlParam.Set("page", strconv.Itoa(page))
	urlParam.Set("per_page", strconv.Itoa(limit))

	path := utils.GenQueryURL(github.vcs.Address, "/user/repos", urlParam)
	response, _, err := githubRequest(path, "GET", github.vcs.VcsToken, nil)
	if err != nil {
		return e.New(e.VcsError, err)
	}

	if response.StatusCode > 300 {
		return e.New(e.VcsInvalidToken, fmt.Sprintf("token valid check response code: %d", response.StatusCode))
	}

	return nil
}

func (v *githubVcs) RepoBaseHttpAddr() string {
	u, err := url.Parse(v.vcs.Address)
	if err != nil {
		return ""
	}

	u.Host = strings.TrimPrefix(u.Host, "api.")

	return u.String()
}

type githubRepoIface struct {
	vcs        *models.Vcs
	repository *RepositoryGithub
	total      int
}

type githubBranch struct {
	Name string `json:"name" form:"name" `
}

func (github *githubRepoIface) ListBranches() ([]string, error) {
	// FIXME github分页默认值最大100，临时处理返回100个
	path := utils.GenQueryURL(github.vcs.Address,
		fmt.Sprintf("/repos/%s/branches?page=1&per_page=100", github.repository.FullName), nil)
	_, body, err := githubRequest(path, "GET", github.vcs.VcsToken, nil)
	if err != nil {
		return nil, e.New(e.VcsError, err)
	}
	rep := make([]githubBranch, 0)

	_ = json.Unmarshal(body, &rep)
	branchList := make([]string, 0)
	for _, v := range rep {
		branchList = append(branchList, v.Name)
	}

	return branchList, nil
}

type githubTag struct {
	Name string `json:"name" form:"name" `
}

func (github *githubRepoIface) ListTags() ([]string, error) {
	// FIXME github分页默认值最大100，临时处理返回100个
	path := utils.GenQueryURL(github.vcs.Address, fmt.Sprintf("/repos/%s/tags?page=1&per_page=100", github.repository.FullName), nil)
	_, body, err := githubRequest(path, "GET", github.vcs.VcsToken, nil)
	if err != nil {
		return nil, e.New(e.VcsError, err)
	}
	rep := make([]githubTag, 0)

	_ = json.Unmarshal(body, &rep)
	tagList := []string{}
	for _, v := range rep {
		tagList = append(tagList, v.Name)
	}
	return tagList, nil

}

type githubCommit struct {
	Sha string `json:"sha"`
}

// BranchCommitId doc: https://docs.github.com/en/rest/reference/repos#get-a-commit
func (github *githubRepoIface) BranchCommitId(branch string) (string, error) {
	path := utils.GenQueryURL(github.vcs.Address,
		fmt.Sprintf("/repos/%s/commits/%s", github.repository.FullName, branch), nil)
	_, body, err := githubRequest(path, "GET", github.vcs.VcsToken, nil)
	if err != nil {
		return "", e.New(e.VcsError, err)
	}

	resp := githubCommit{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return "", e.New(e.VcsError, err)
	}
	if resp.Sha == "" {
		logs.Get().Warnf("query github branch commit it failed")
		return "", e.New(e.VcsError, fmt.Errorf("query commit id failed"))
	}
	return resp.Sha, nil
}

type githubFiles struct {
	Type string `json:"type" form:"type" `
	Path string `json:"path" form:"path" `
	Name string `json:"name" form:"name" `
}

func (github *githubRepoIface) ListFiles(option VcsIfaceOptions) ([]string, error) {
	var (
		ansible = "ansible"
	)
	urlParam := url.Values{}
	urlParam.Set("ref", getBranch(github, option.Ref))
	var path string
	if option.Path != "" {
		path = utils.GenQueryURL(github.vcs.Address,
			fmt.Sprintf("/repos/%s/contents/%s", github.repository.FullName, option.Path), urlParam)
	} else {
		path = utils.GenQueryURL(github.vcs.Address,
			fmt.Sprintf("/repos/%s/contents", github.repository.FullName), urlParam)
	}
	_, body, er := githubRequest(path, "GET", github.vcs.VcsToken, nil)
	if er != nil {
		return []string{}, e.New(e.VcsError, er)
	}
	resp := make([]string, 0)
	rep := make([]githubFiles, 0)
	resp, err := github.UpdatePlaybookWorkDir(resp, body, option, ansible)
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(body, &rep)
	for _, v := range rep {
		if v.Type == "dir" && option.Recursive {
			option.Path = v.Path
			repList, _ := github.ListFiles(option)
			resp = append(resp, repList...)
		}

		if v.Type == "file" && matchGlob(option.Search, v.Name) {
			resp = append(resp, v.Path)
		}

	}

	return resp, nil

}

type githubReadContent struct {
	Content string `json:"content" form:"content" `
}

type githubReadTarget struct {
	Target string `json:"target" form:"target" `
}

func (github *githubRepoIface) UpdatePlaybookWorkDir(resp []string, body []byte, option VcsIfaceOptions, pattern string) ([]string, error) {
	if strings.Contains(option.Path, pattern) {
		gf := githubFiles{}
		json.Unmarshal(body, &gf)
		if gf.Type == "symlink" && gf.Name == pattern {
			grt := githubReadTarget{}
			if err := json.Unmarshal(body, &grt); err != nil {
				return resp, err
			}
			Path := strings.Replace(option.Path, pattern, "", 1)
			Paths := filepath.Join(Path, grt.Target)
			option.Path = Paths
			repList, _ := github.ListFiles(option)
			resp = append(resp, repList...)
			return resp, nil
		}
	}
	return resp, nil
}

func (github *githubRepoIface) ReadFileContent(branch, path string) (content []byte, err error) {
	defer func() {
		if err != nil && strings.Contains(err.Error(), "Not Found") {
			err = e.New(e.ObjectNotExists)
		}
	}()

	urlParam := url.Values{}
	urlParam.Set("ref", branch)
	pathAddr := utils.GenQueryURL(github.vcs.Address,
		fmt.Sprintf("/repos/%s/contents/%s", github.repository.FullName, path), urlParam)
	response, body, er := githubRequest(pathAddr, "GET", github.vcs.VcsToken, nil)
	if er != nil {
		return nil, e.New(e.VcsError, er)
	}
	grc := githubReadContent{}
	if err := json.Unmarshal(body[:], &grc); err != nil {
		return nil, err
	}

	if response.StatusCode >= 300 {
		err = e.New(e.VcsError, fmt.Errorf("%s: %s", response.Status, body))
		return []byte{}, err
	}

	decoded, err := base64.StdEncoding.DecodeString(grc.Content)
	if err != nil {
		return nil, e.New(e.VcsError, er)
	}
	return decoded[:], nil

}

func (github *githubRepoIface) FormatRepoSearch() (project *Projects, err e.Error) {
	return &Projects{
		ID:             github.repository.FullName,
		Description:    github.repository.Description,
		DefaultBranch:  github.repository.DefaultBranch,
		SSHURLToRepo:   github.repository.SSHURL,
		HTTPURLToRepo:  github.repository.CloneURL,
		Name:           github.repository.Name,
		LastActivityAt: &github.repository.Updated,
		FullName:       github.repository.FullName,
	}, nil
}

func (github *githubRepoIface) DefaultBranch() string {
	return github.repository.DefaultBranch
}

// AddWebhook doc: https://docs.github.com/cn/rest/reference/repos#traffic
func (github *githubRepoIface) AddWebhook(url string) error {
	path := utils.GenQueryURL(github.vcs.Address, fmt.Sprintf("/repos/%s/hooks", github.repository.FullName), nil)
	bodys := map[string]interface{}{
		"config": map[string]interface{}{
			"url":          url,
			"content_type": "json",
		},
		"events": []string{
			"pull_request",
			"push",
		},
		"type":   "gitea",
		"active": true,
	}
	b, _ := json.Marshal(&bodys)
	response, respBody, err := githubRequest(path, "POST", github.vcs.VcsToken, b)

	if err != nil {
		return e.New(e.VcsError, err)
	}

	if response.StatusCode >= 300 && !strings.Contains(string(respBody), "Hook already exists on this repository") {
		err = e.New(e.VcsError, fmt.Errorf("%s: %s", response.Status, string(respBody)))
		return err
	}
	return nil
}

func (github *githubRepoIface) ListWebhook() ([]RepoHook, error) {
	path := utils.GenQueryURL(github.vcs.Address, fmt.Sprintf("/repos/%s/hooks", github.repository.FullName), nil)
	_, body, err := githubRequest(path, "GET", github.vcs.VcsToken, nil)
	if err != nil {
		return nil, e.New(e.VcsError, err)
	}

	return initRepoHook(body), nil
}

func (github *githubRepoIface) DeleteWebhook(id int) error {
	path := utils.GenQueryURL(github.vcs.Address, fmt.Sprintf("/repos/%s/hooks/%d", github.repository.FullName, id), nil)
	_, _, err := githubRequest(path, "DELETE", github.vcs.VcsToken, nil)
	if err != nil {
		return e.New(e.VcsError, err)
	}
	return nil
}

// CreatePrComment doc: https://docs.github.com/en/rest/reference/pulls#submit-a-review-for-a-pull-request
func (github *githubRepoIface) CreatePrComment(prId int, comment string) error {
	path := utils.GenQueryURL(github.vcs.Address, fmt.Sprintf("/repos/%s/pulls/%d/reviews", github.repository.FullName, prId), nil)
	requestBody := map[string]string{
		"body":  comment,
		"event": "COMMENT",
	}
	b, er := json.Marshal(requestBody)
	if er != nil {
		return er
	}
	response, body, err := githubRequest(path, http.MethodPost, github.vcs.VcsToken, b)

	if err != nil {
		return e.New(e.VcsError, err)
	}

	if response.StatusCode > 300 {
		return e.New(e.VcsError, fmt.Errorf("code: %s, err: %s", response.Status, string(body)))
	}
	return nil
}

func (github *githubRepoIface) GetFullFilePath(address, filePath, repoRevision string) string {
	u, _ := url.Parse("https://github.com/")
	u.Path = path.Join(u.Path, github.repository.FullName, "blob", repoRevision, filePath)
	return u.String()
}

func (github *githubRepoIface) GetCommitFullPath(address, commitId string) string {
	u, _ := url.Parse("https://github.com/")
	u.Path = path.Join(u.Path, github.repository.FullName, "commit", commitId)
	return u.String()
}

// giteaRequest
// param path : gitea api路径
// param method 请求方式
func githubRequest(path, method, token string, requestBody []byte) (*http.Response, []byte, error) {
	vcsToken, err := GetVcsToken(token)
	if err != nil {
		return nil, nil, err
	}

	request, er := http.NewRequest(method, path, bytes.NewBuffer(requestBody))
	if er != nil {
		return nil, nil, er
	}
	client := &http.Client{}
	request.Header.Set("Content-Type", "multipart/form-data")
	request.Header.Set("Accept", "application/vnd.github.v3+json")
	request.Header.Set("Authorization", fmt.Sprintf("token %s", vcsToken))
	response, err := client.Do(request)
	if err != nil {
		return nil, nil, err
	}
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)

	return response, body, nil

}
