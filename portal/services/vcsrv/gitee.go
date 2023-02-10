// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package vcsrv

import (
	"bytes"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/models"
	"cloudiac/utils"
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
)

// newGiteeInstance
// gitee open api文档: https://gitee.com/api/v5/swagger#/getV5ReposOwnerRepoBranches
func newGiteeInstance(vcs *models.Vcs) (VcsIface, error) {
	vcs.Address = fmt.Sprintf("%s/api/v5", utils.GetUrl(vcs.Address))
	vcsToken, err := vcs.DecryptToken()
	if err != nil {
		return nil, err
	}
	param := url.Values{}
	param.Add("access_token", vcsToken)
	return &giteeVcs{vcs: vcs, urlParam: param}, nil
}

type giteeVcs struct {
	vcs      *models.Vcs
	urlParam url.Values
}

func (gitee *giteeVcs) GetRepo(idOrPath string) (RepoIface, error) {
	path := gitee.vcs.Address + fmt.Sprintf("/repos/%s?access_token=%s", idOrPath, gitee.urlParam.Get("access_token"))
	_, body, er := giteeRequest(path, "GET", nil)
	if er != nil {
		return nil, e.New(e.VcsError, er)
	}

	rep := RepositoryGitee{}
	_ = json.Unmarshal(body, &rep)
	return &giteeRepoIface{
		vcs:        gitee.vcs,
		repository: &rep,
		urlParam:   gitee.urlParam,
	}, nil
}

func (v *giteeVcs) RepoBaseHttpAddr() string {
	return strings.TrimSuffix(v.vcs.Address, "/api/v5")
}

type RepositoryGitee struct {
	ID            int64     `json:"id"`
	Description   string    `json:"description"`
	DefaultBranch string    `json:"default_branch"`
	HtmlUrl       string    `json:"html_url"`
	SSHURL        string    `json:"ssh_url"`
	Name          string    `json:"name"`
	Updated       time.Time `json:"updated_at"`
	FullName      string    `json:"full_name" form:"full_name" `
}

func (gitee *giteeVcs) ListRepos(namespace, search string, limit, offset int) ([]RepoIface, int64, error) {
	link, _ := url.Parse("/user/repos")
	page := utils.LimitOffset2Page(limit, offset)
	link.RawQuery += fmt.Sprintf("access_token=%s&page=%d&per_page=%d", gitee.urlParam.Get("access_token"), page, limit)
	if search != "" {
		link.RawQuery += fmt.Sprintf("&q=%s", search)
	}
	path := gitee.vcs.Address + link.String()
	response, body, err := giteeRequest(path, "GET", nil)

	if err != nil {
		return nil, 0, e.New(e.VcsError, err)
	}

	var total int64
	if len(response.Header["Total_count"]) != 0 {
		total, _ = strconv.ParseInt(response.Header["Total_count"][0], 10, 64)
	}
	rep := make([]RepositoryGitee, 0)
	_ = json.Unmarshal(body, &rep)

	repoList := make([]RepoIface, 0)
	for index := range rep {
		repoList = append(repoList, &giteeRepoIface{
			vcs:        gitee.vcs,
			repository: &rep[index],
			urlParam:   gitee.urlParam,
		})
	}

	return repoList, total, nil
}

// https://gitee.com/api/v5/user
func (gitee *giteeVcs) UserInfo() (UserInfo, error) {
	path := gitee.vcs.Address + fmt.Sprintf("/user?access_token=%s", gitee.urlParam.Get("access_token"))
	_, body, er := giteeRequest(path, "GET", nil)
	if er != nil {
		return UserInfo{}, e.New(e.VcsError, er)
	}

	rep := UserInfo{}
	_ = json.Unmarshal(body, &rep)
	return rep, nil
}

func (gitee *giteeVcs) TokenCheck() error {
	limit, offset := 1, 1
	link, _ := url.Parse("/user/repos")
	page := utils.LimitOffset2Page(limit, offset)
	link.RawQuery += fmt.Sprintf("access_token=%s&page=%d&per_page=%d",
		gitee.urlParam.Get("access_token"), page, limit)
	path := gitee.vcs.Address + link.String()
	response, _, err := giteeRequest(path, "GET", nil)
	if err != nil {
		return e.New(e.VcsError, err)
	}

	if response.StatusCode > 300 {
		return e.New(e.VcsInvalidToken, fmt.Sprintf("token valid check response code: %d", response.StatusCode))
	}

	return nil
}

type giteeRepoIface struct {
	vcs        *models.Vcs
	repository *RepositoryGitee
	urlParam   url.Values
}

type giteeBranch struct {
	Name string `json:"name" form:"name" `
}

func (gitee *giteeRepoIface) ListBranches() ([]string, error) {
	path := gitee.vcs.Address +
		fmt.Sprintf("/repos/%s/branches?access_token=%s", gitee.repository.FullName, gitee.urlParam.Get("access_token"))
	_, body, err := giteeRequest(path, "GET", nil)
	if err != nil {
		return nil, e.New(e.VcsError, err)
	}

	rep := make([]giteeBranch, 0)

	_ = json.Unmarshal(body, &rep)
	branchList := []string{}
	for _, v := range rep {
		branchList = append(branchList, v.Name)
	}
	return branchList, nil
}

type giteeTag struct {
	Name string `json:"name" form:"name" `
}

func (gitee *giteeRepoIface) ListTags() ([]string, error) {
	path := gitee.vcs.Address + fmt.Sprintf("/repos/%s/tags", gitee.repository.FullName)
	_, body, err := giteeRequest(path, "GET", nil)
	if err != nil {
		return nil, e.New(e.VcsError, err)
	}

	rep := make([]giteeTag, 0)

	_ = json.Unmarshal(body, &rep)
	tagList := []string{}
	for _, v := range rep {
		tagList = append(tagList, v.Name)
	}
	return tagList, nil

}

type giteeCommit struct {
	CommitId string `json:"sha" form:"sha" `
}

func (gitee *giteeRepoIface) BranchCommitId(branch string) (string, error) {
	path := gitee.vcs.Address +
		fmt.Sprintf("/repos/%s/commits/%s?access_token=%s", gitee.repository.FullName, branch, gitee.urlParam.Get("access_token"))
	_, body, err := giteeRequest(path, "GET", nil)
	if err != nil {
		return "", e.New(e.VcsError, err)
	}

	rep := giteeCommit{}
	_ = json.Unmarshal(body, &rep)
	return rep.CommitId, nil
}

type giteeFiles struct {
	Type string `json:"type" form:"type" `
	Path string `json:"path" form:"path" `
	Name string `json:"name" form:"name" `
}

func (gitee *giteeRepoIface) ListFiles(option VcsIfaceOptions) ([]string, error) {
	var (
		ansible = "ansible"
	)
	var path string = gitee.vcs.Address
	branch := getBranch(gitee, option.Ref)
	if option.Path != "" {
		path += fmt.Sprintf("/repos/%s/contents/%s?access_token=%s&ref=%s", //nolint
			gitee.repository.FullName, option.Path, gitee.urlParam.Get("access_token"), branch)
	} else {
		path += fmt.Sprintf("/repos/%s/contents/%s?access_token=%s&ref=%s", //nolint
			gitee.repository.FullName, "%2F", gitee.urlParam.Get("access_token"), branch)
	}
	_, body, er := giteeRequest(path, "GET", nil)
	if er != nil {
		return []string{}, e.New(e.VcsError, er)
	}

	resp := make([]string, 0)
	rep := make([]giteeFiles, 0)
	resp, err := gitee.UpdatePlaybookWorkDir(resp, body, option, ansible)
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(body, &rep)
	for _, v := range rep {
		if v.Type == "dir" && option.Recursive {
			option.Path = v.Path
			repList, _ := gitee.ListFiles(option)
			resp = append(resp, repList...)
		}

		if v.Type == "file" && matchGlob(option.Search, v.Name) {
			resp = append(resp, v.Path)
		}

	}

	return resp, nil
}

type giteeReadContent struct {
	Content string `json:"content" form:"content" `
}

func (gitee *giteeRepoIface) UpdatePlaybookWorkDir(resp []string, body []byte, option VcsIfaceOptions, pattern string) ([]string, error) {
	if strings.Contains(option.Path, pattern) {
		gf := giteeFiles{}
		json.Unmarshal(body, &gf)
		if gf.Type == "symlink" && gf.Name == pattern {
			content, err := gitee.ReadFileContent(getBranch(gitee, option.Ref), option.Path)
			if err != nil {
				return resp, err
			}
			Path := strings.Replace(option.Path, pattern, "", 1)
			Paths := filepath.Join(Path, string(content))
			option.Path = Paths
			repList, _ := gitee.ListFiles(option)
			resp = append(resp, repList...)
			return resp, nil
		}
	}
	return resp, nil
}

func (gitee *giteeRepoIface) ReadFileContent(branch, path string) (content []byte, err error) {
	pathAddr := gitee.vcs.Address +
		fmt.Sprintf("/repos/%s/contents/%s?access_token=%s&ref=%s", gitee.repository.FullName, path, gitee.urlParam.Get("access_token"), branch)
	_, body, er := giteeRequest(pathAddr, "GET", nil) //nolint

	if er != nil {
		return nil, e.New(e.VcsError, er)
	}

	grc := giteeReadContent{}
	if err := json.Unmarshal(body[:], &grc); err != nil {
		// 找不到文件时状态码为200，gieee接口会返回'[]'
		if string(body) == "[]" {
			return nil, e.New(e.ObjectNotExists)
		}
		return nil, fmt.Errorf("json unmarshl err: %v, body: %s", err, string(body))
	}

	decoded, err := base64.StdEncoding.DecodeString(grc.Content)
	if err != nil {
		return nil, e.New(e.VcsError, er)
	}
	return decoded[:], nil
}

func (gitee *giteeRepoIface) FormatRepoSearch() (project *Projects, err e.Error) {
	return &Projects{
		ID:             gitee.repository.FullName,
		Description:    gitee.repository.Description,
		DefaultBranch:  gitee.repository.DefaultBranch,
		SSHURLToRepo:   gitee.repository.SSHURL,
		HTTPURLToRepo:  gitee.repository.HtmlUrl,
		Name:           gitee.repository.Name,
		LastActivityAt: &gitee.repository.Updated,
		FullName:       gitee.repository.FullName,
	}, nil
}

func (gitee *giteeRepoIface) DefaultBranch() string {
	return gitee.repository.DefaultBranch
}

// AddWebhook doc: https://gitee.com/api/v5/swagger#/deleteV5ReposOwnerRepoHooksId
func (gitee *giteeRepoIface) AddWebhook(url string) error {
	path := gitee.vcs.Address +
		fmt.Sprintf("/repos/%s/hooks?access_token=%s", gitee.repository.FullName, gitee.urlParam.Get("access_token"))
	body := map[string]interface{}{
		"url":                   url,
		"push_events":           "true",
		"merge_requests_events": "true",
	}
	b, _ := json.Marshal(&body)
	response, respBody, err := giteeRequest(path, http.MethodPost, b)

	if err != nil {
		return e.New(e.VcsError, err)
	}

	if response.StatusCode >= 300 {
		err = e.New(e.VcsError, fmt.Errorf("%s: %s", response.Status, string(respBody)))
		return err
	}
	return nil
}

func initGiteeRepoHook(body []byte) []RepoHook {
	ph := make([]struct {
		Url string `json:"url"`
		Id  int    `json:"id"`
	}, 0)
	_ = json.Unmarshal(body, &ph)

	resp := make([]RepoHook, 0)

	for _, v := range ph {
		resp = append(resp, RepoHook{
			Id:  v.Id,
			Url: v.Url,
		})
	}
	return resp
}

func (gitee *giteeRepoIface) ListWebhook() ([]RepoHook, error) {
	path := gitee.vcs.Address +
		fmt.Sprintf("/repos/%s/hooks?access_token=%s", gitee.repository.FullName, gitee.urlParam.Get("access_token"))
	_, body, err := giteeRequest(path, http.MethodGet, nil)
	if err != nil {
		return nil, e.New(e.VcsError, err)
	}

	return initGiteeRepoHook(body), nil
}

func (gitee *giteeRepoIface) DeleteWebhook(id int) error {
	path := gitee.vcs.Address +
		fmt.Sprintf("/repos/%s/hooks/%d?access_token=%s", gitee.repository.FullName, id, gitee.urlParam.Get("access_token"))
	_, _, err := giteeRequest(path, "DELETE", nil)
	if err != nil {
		return e.New(e.VcsError, err)
	}
	return nil
}

func (gitee *giteeRepoIface) CreatePrComment(prId int, comment string) error {
	path := gitee.vcs.Address +
		fmt.Sprintf("/repos/%s/pulls/%d/comments?access_token=%s", gitee.repository.FullName, prId, gitee.urlParam.Get("access_token"))

	requestBody := map[string]string{
		"body": comment,
	}
	b, er := json.Marshal(requestBody)
	if er != nil {
		return er
	}
	_, _, err := giteeRequest(path, http.MethodPost, b)
	if err != nil {
		return e.New(e.VcsError, err)
	}
	return nil
}

func (gitee *giteeRepoIface) GetFullFilePath(address, filePath, repoRevision string) string {
	u, _ := url.Parse(address)
	u.Path = path.Join(u.Path, gitee.repository.FullName, "blob", repoRevision, filePath)
	return u.String()
}

func (gitee *giteeRepoIface) GetCommitFullPath(address, commitId string) string {
	u, _ := url.Parse(address)
	u.Path = path.Join(u.Path, gitee.repository.FullName, "commit", commitId)
	return u.String()
}

// giteeRequest
// param path : gitea api路径
// param method 请求方式
func giteeRequest(path, method string, requestBody []byte) (*http.Response, []byte, error) {
	request, er := http.NewRequest(method, path, bytes.NewBuffer(requestBody))
	if er != nil {
		return nil, nil, er
	}
	client := &http.Client{}
	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		return nil, nil, err
	}
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	return response, body, nil

}
