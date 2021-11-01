// Copyright 2021 CloudJ Company Limited. All rights reserved.

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
	"strconv"
	"strings"
	"time"
)

//newGiteeInstance
//gitee open api文档: https://gitee.com/api/v5/swagger#/getV5ReposOwnerRepoBranches
func newGiteeInstance(vcs *models.Vcs) (VcsIface, error) {
	vcs.Address = fmt.Sprintf("%s/api/v5", utils.GetUrl(vcs.Address))
	return &giteeVcs{vcs: vcs}, nil
}

type giteeVcs struct {
	vcs *models.Vcs
}

func (gitee *giteeVcs) GetRepo(idOrPath string) (RepoIface, error) {
	path := gitee.vcs.Address + fmt.Sprintf("/repos/%s?access_token=%s", idOrPath, gitee.vcs.VcsToken)
	_, body, er := giteeRequest(path, "GET", nil)
	if er != nil {
		return nil, e.New(e.BadRequest, er)
	}

	rep := RepositoryGitee{}
	_ = json.Unmarshal(body, &rep)
	return &giteeRepoIface{
		vcs:        gitee.vcs,
		repository: &rep,
	}, nil

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
	link.RawQuery += fmt.Sprintf("access_token=%s&page=%d&per_page=%d", gitee.vcs.VcsToken, page, limit)
	if search != "" {
		link.RawQuery += fmt.Sprintf("&q=%s", search)
	}
	path := gitee.vcs.Address + link.String()
	response, body, err := giteeRequest(path, "GET", nil)

	if err != nil {
		return nil, 0, e.New(e.BadRequest, err)
	}

	var total int64
	if len(response.Header["Total_count"]) != 0 {
		total, _ = strconv.ParseInt(response.Header["Total_count"][0], 10, 64)
	}
	rep := make([]RepositoryGitee, 0)
	_ = json.Unmarshal(body, &rep)

	repoList := make([]RepoIface, 0)
	for index, _ := range rep {
		repoList = append(repoList, &giteeRepoIface{
			vcs:        gitee.vcs,
			repository: &rep[index],
		})
	}

	return repoList, total, nil
}

type giteeRepoIface struct {
	vcs        *models.Vcs
	repository *RepositoryGitee
}

type giteeBranch struct {
	Name string `json:"name" form:"name" `
}

func (gitee *giteeRepoIface) ListBranches() ([]string, error) {
	path := gitee.vcs.Address +
		fmt.Sprintf("/repos/%s/branches?access_token=%s", gitee.repository.FullName, gitee.vcs.VcsToken)
	_, body, err := giteeRequest(path, "GET", nil)
	if err != nil {
		return nil, e.New(e.BadRequest, err)
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
		return nil, e.New(e.BadRequest, err)
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
		fmt.Sprintf("/repos/%s/commits/%s?access_token=%s", gitee.repository.FullName, branch, gitee.vcs.VcsToken)
	_, body, err := giteeRequest(path, "GET", nil)
	if err != nil {
		return "", e.New(e.BadRequest, err)
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
	var path string = gitee.vcs.Address
	branch := getBranch(gitee, option.Ref)
	if option.Path != "" {
		path += fmt.Sprintf("/repos/%s/contents/%s?access_token=%s&ref=%s",
			gitee.repository.FullName, option.Path, gitee.vcs.VcsToken, branch)
	} else {
		path += fmt.Sprintf("/repos/%s/contents/%s?access_token=%s&ref=%s",
			gitee.repository.FullName, "%2F", gitee.vcs.VcsToken, branch)
	}
	_, body, er := giteeRequest(path, "GET", nil)
	if er != nil {
		return []string{}, e.New(e.BadRequest, er)
	}

	resp := make([]string, 0)
	rep := make([]giteeFiles, 0)
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

func (gitee *giteeRepoIface) ReadFileContent(branch, path string) (content []byte, err error) {
	defer func() {
		if err != nil && (strings.Contains(err.Error(), "Not Found") ||
			strings.Contains(err.Error(), "json: cannot unmarshal")) {
			err = e.New(e.ObjectNotExists)
		}
	}()
	pathAddr := gitee.vcs.Address +
		fmt.Sprintf("/repos/%s/contents/%s?access_token=%s&ref=%s", gitee.repository.FullName, path, gitee.vcs.VcsToken, branch)
	response, body, er := giteeRequest(pathAddr, "GET", nil)

	if er != nil {
		return nil, e.New(e.BadRequest, er)
	}

	grc := giteeReadContent{}
	if err := json.Unmarshal(body[:], &grc); err != nil {
		return nil, err
	}

	if response.StatusCode >= 300 {
		err = e.New(e.VcsError, fmt.Errorf("%s: %s", response.Status, body))
		return []byte{}, err
	}

	decoded, err := base64.StdEncoding.DecodeString(grc.Content)
	if err != nil {
		return nil, e.New(e.BadRequest, er)
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

//AddWebhook doc: https://gitee.com/api/v5/swagger#/deleteV5ReposOwnerRepoHooksId
func (gitee *giteeRepoIface) AddWebhook(url string) error {
	path := gitee.vcs.Address +
		fmt.Sprintf("/repos/%s/hooks?access_token=%s", gitee.repository.FullName, gitee.vcs.VcsToken)
	body := map[string]interface{}{
		"url":                   url,
		"push_events":           "true",
		"merge_requests_events": "true",
	}
	b, _ := json.Marshal(&body)
	_, _, err := giteeRequest(path, "POST", b)
	if err != nil {
		return e.New(e.BadRequest, err)
	}
	return nil
}

func (gitee *giteeRepoIface) ListWebhook() ([]ProjectsHook, error) {
	ph := make([]ProjectsHook, 0)
	path := gitee.vcs.Address +
		fmt.Sprintf("/repos/%s/hooks?access_token=%s", gitee.repository.FullName, gitee.vcs.VcsToken)
	_, body, err := giteeRequest(path, "GET", nil)
	if err != nil {
		return ph, e.New(e.BadRequest, err)
	}

	rep := []giteeCommit{}
	_ = json.Unmarshal(body, &rep)

	return ph, nil
}

func (gitee *giteeRepoIface) DeleteWebhook(id int) error {
	path := gitee.vcs.Address +
		fmt.Sprintf("/repos/%s/hooks/%d?access_token=%s", gitee.repository.FullName, id, gitee.vcs.VcsToken)
	_, _, err := giteeRequest(path, "DELETE", nil)
	if err != nil {
		return e.New(e.BadRequest, err)
	}
	return nil
}

//giteeRequest
//param path : gitea api路径
//param method 请求方式
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
