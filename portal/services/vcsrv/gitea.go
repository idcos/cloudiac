// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package vcsrv

import (
	"bytes"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/models"
	"cloudiac/utils"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func newGiteaInstance(vcs *models.Vcs) (VcsIface, error) {
	vcs.Address = utils.GetUrl(vcs.Address)
	return &giteaVcs{vcs: vcs}, nil

}

type giteaVcs struct {
	vcs *models.Vcs
}

const giteaApiRoute = "/api/v1"

func (gitea *giteaVcs) GetRepo(idOrPath string) (RepoIface, error) {
	path := gitea.vcs.Address + fmt.Sprintf("%s/repositories/%s", giteaApiRoute, idOrPath)
	_, body, er := giteaRequest(path, "GET", gitea.vcs.VcsToken, nil)
	if er != nil {
		return nil, e.New(e.BadRequest, er)
	}
	rep := Repository{}
	_ = json.Unmarshal(body, &rep)
	return &giteaRepoIface{
		vcs:        gitea.vcs,
		repository: &rep,
	}, nil

}

type SearchRepoResponse struct {
	Repos []*Repository `json:"data"`
}

type Repository struct {
	ID            int64     `json:"id"`
	Description   string    `json:"description"`
	DefaultBranch string    `json:"default_branch"`
	SSHURL        string    `json:"ssh_url"`
	CloneURL      string    `json:"clone_url"`
	Name          string    `json:"name"`
	Updated       time.Time `json:"updated_at"`
	FullName      string    `json:"full_name" form:"full_name" `
}

/*curl -X 'GET' \
'http://localhost:9999/api/v1/repos/search?page=1&limit=1&exclusive=true&uid=1&q=test' \
-H 'accept: application/json' \
-H 'Authorization: token 27b9b370eb3qqqqqqqqc0f3a5de10a3'
*/

//ListRepos Fixme中的数据不能直接调用repo接口的方法
func (gitea *giteaVcs) ListRepos(namespace, search string, limit, offset int) ([]RepoIface, int64, error) {
	user, err := getGiteaUserMe(gitea.vcs)
	if err != nil {
		return nil, 0, e.New(e.BadRequest, err)
	}
	link, _ := url.Parse("/repos/search")
	page := utils.LimitOffset2Page(limit, offset)
	link.RawQuery = fmt.Sprintf("page=%d&limit=%d&exclusive=true&uid=%d", page, limit, user.Id)
	if search != "" {
		link.RawQuery = link.RawQuery + fmt.Sprintf("&q=%s", search)
	}
	path := gitea.vcs.Address + giteaApiRoute + link.String()
	response, body, err := giteaRequest(path, "GET", gitea.vcs.VcsToken, nil)

	if err != nil {
		return nil, 0, e.New(e.BadRequest, err)
	}

	var total int64
	if len(response.Header["X-Total-Count"]) != 0 {
		total, _ = strconv.ParseInt(response.Header["X-Total-Count"][0], 10, 64)
	}

	rep := SearchRepoResponse{}
	_ = json.Unmarshal(body, &rep)

	repoList := make([]RepoIface, 0)
	for _, v := range rep.Repos {
		repoList = append(repoList, &giteaRepoIface{
			vcs:        gitea.vcs,
			repository: v,
		})
	}

	return repoList, total, nil
}

func (gitea *giteaVcs) UserInfo() (UserInfo, error) {

	return UserInfo{}, nil
}

type giteaRepoIface struct {
	vcs        *models.Vcs
	repository *Repository
}

type giteaBranch struct {
	Name string `json:"name" form:"name" `
}

func (gitea *giteaRepoIface) ListBranches() ([]string, error) {
	path := gitea.vcs.Address + giteaApiRoute +
		fmt.Sprintf("/repos/%s/branches?limit=0&page=0", gitea.repository.FullName)

	_, body, err := giteaRequest(path, "GET", gitea.vcs.VcsToken, nil)
	if err != nil {
		return nil, e.New(e.BadRequest, err)
	}
	rep := make([]giteaBranch, 0)

	_ = json.Unmarshal(body, &rep)
	branchList := []string{}
	for _, v := range rep {
		branchList = append(branchList, v.Name)
	}
	return branchList, nil
}

type giteaTag struct {
	Name string `json:"name" form:"name" `
}

func (gitea *giteaRepoIface) ListTags() ([]string, error) {
	path := gitea.vcs.Address + giteaApiRoute + fmt.Sprintf("/repos/%s/tags", gitea.repository.FullName)
	_, body, err := giteaRequest(path, "GET", gitea.vcs.VcsToken, nil)
	if err != nil {
		return nil, e.New(e.BadRequest, err)
	}
	rep := make([]giteaTag, 0)

	_ = json.Unmarshal(body, &rep)
	tagList := []string{}
	for _, v := range rep {
		tagList = append(tagList, v.Name)
	}
	return tagList, nil
}

type giteaCommit struct {
	Commit struct {
		Id string `json:"id" form:"id" `
	} `json:"commit" form:"commit" `
}

func (gitea *giteaRepoIface) BranchCommitId(branch string) (string, error) {
	path := gitea.vcs.Address + giteaApiRoute +
		fmt.Sprintf("/repos/%s/branches/%s?limit=0&page=0", gitea.repository.FullName, branch)
	_, body, err := giteaRequest(path, "GET", gitea.vcs.VcsToken, nil)
	if err != nil {
		return "", e.New(e.BadRequest, err)
	}
	rep := giteaCommit{}
	_ = json.Unmarshal(body, &rep)
	return rep.Commit.Id, nil
}

type giteaFiles struct {
	Type string `json:"type" form:"type" `
	Path string `json:"path" form:"path" `
	Name string `json:"name" form:"name" `
}

func (gitea *giteaRepoIface) ListFiles(option VcsIfaceOptions) ([]string, error) {
	var path string = gitea.vcs.Address
	branch := getBranch(gitea, option.Ref)
	if option.Path != "" {
		path += giteaApiRoute +
			fmt.Sprintf("/repos/%s/contents/%s?limit=0&page=0&ref=%s",
				gitea.repository.FullName, option.Path, branch)
	} else {
		path += giteaApiRoute +
			fmt.Sprintf("/repos/%s/contents?limit=0&page=0&ref=%s",
				gitea.repository.FullName, branch)
	}
	_, body, er := giteaRequest(path, "GET", gitea.vcs.VcsToken, nil)
	if er != nil {
		return []string{}, e.New(e.BadRequest, er)
	}
	resp := make([]string, 0)
	rep := make([]giteaFiles, 0)
	_ = json.Unmarshal(body, &rep)
	for _, v := range rep {
		if v.Type == "dir" && option.Recursive {
			option.Path = v.Path
			repList, _ := gitea.ListFiles(option)
			resp = append(resp, repList...)
		}
		if v.Type == "file" && matchGlob(option.Search, v.Name) {
			resp = append(resp, v.Path)
		}

	}

	return resp, nil
}

func (gitea *giteaRepoIface) ReadFileContent(branch, path string) (content []byte, err error) {
	defer func() {
		if err != nil && strings.Contains(err.Error(), "Not Found") {
			err = e.New(e.ObjectNotExists)
		}
	}()

	pathAddr := gitea.vcs.Address + giteaApiRoute +
		fmt.Sprintf("/repos/%s/raw/%s?ref=%s", gitea.repository.FullName, path, branch)
	response, body, er := giteaRequest(pathAddr, "GET", gitea.vcs.VcsToken, nil)
	if er != nil {
		return []byte{}, e.New(e.BadRequest, er)
	}
	defer response.Body.Close()

	if response.StatusCode >= 300 {
		err = e.New(e.VcsError, fmt.Errorf("%s: %s", response.Status, body))
		return []byte{}, err
	}

	return body[:], nil
}

func (gitea *giteaRepoIface) FormatRepoSearch() (project *Projects, err e.Error) {
	return &Projects{
		ID:             fmt.Sprintf("%d", gitea.repository.ID),
		Description:    gitea.repository.Description,
		DefaultBranch:  gitea.repository.DefaultBranch,
		SSHURLToRepo:   gitea.repository.SSHURL,
		HTTPURLToRepo:  gitea.repository.CloneURL,
		Name:           gitea.repository.Name,
		LastActivityAt: &gitea.repository.Updated,
		FullName:       gitea.repository.FullName,
	}, nil
}

func (gitea *giteaRepoIface) DefaultBranch() string {
	return gitea.repository.DefaultBranch
}

//AddWebhook doc: http://10.0.3.124:3000/api/swagger#/repository/repoDeleteHook
func (gitea *giteaRepoIface) AddWebhook(url string) error {
	path := gitea.vcs.Address + giteaApiRoute + fmt.Sprintf("/repos/%s/hooks", gitea.repository.FullName)
	bodys := map[string]interface{}{
		"active": true,
		"config": map[string]interface{}{
			"url":          url,
			"content_type": "json",
		},
		"events": []string{
			"pull_request_only",
			"push",
		},
		"type": "gitea",
	}
	b, _ := json.Marshal(&bodys)
	response, body, err := giteaRequest(path, http.MethodPost, gitea.vcs.VcsToken, b)
	if err != nil {
		return e.New(e.BadRequest, err)
	}

	if response.StatusCode >= 300 {
		err = e.New(e.VcsError, fmt.Errorf("%s: %s", response.Status, string(body)))
		return err
	}

	return nil
}

func (gitea *giteaRepoIface) ListWebhook() ([]RepoHook, error) {
	path := gitea.vcs.Address + giteaApiRoute + fmt.Sprintf("/repos/%s/hooks", gitea.repository.FullName)
	_, body, err := giteaRequest(path, "GET", gitea.vcs.VcsToken, nil)
	if err != nil {
		return nil, e.New(e.BadRequest, err)
	}

	ph := make([]struct {
		Config struct {
			ContentType string `json:"content_type"`
			Url         string `json:"url"`
		} `json:"config"`
		Id int `json:"id"`
	}, 0)
	_ = json.Unmarshal(body, &ph)

	resp := make([]RepoHook, 0)
	for _, v := range ph {
		resp = append(resp, RepoHook{
			Id:  v.Id,
			Url: v.Config.Url,
		})
	}
	return resp, nil
}

func (gitea *giteaRepoIface) DeleteWebhook(id int) error {
	path := gitea.vcs.Address + giteaApiRoute + fmt.Sprintf("/repos/%s/hooks/%d", gitea.repository.FullName, id)
	_, body, err := giteaRequest(path, "DELETE", gitea.vcs.VcsToken, nil)
	if err != nil {
		return e.New(e.BadRequest, err)
	}
	rep := make([]giteaTag, 0)

	_ = json.Unmarshal(body, &rep)
	return nil
}

func (gitea *giteaRepoIface) CreatePrComment(prId int, comment string) error {
	path := gitea.vcs.Address + giteaApiRoute + fmt.Sprintf("/repos/%s/pulls/%d/reviews", gitea.repository.FullName, prId)
	requestBody := map[string]string{
		"body": comment,
	}
	b, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}
	_, _, err = giteaRequest(path, http.MethodPost, gitea.vcs.VcsToken, b)
	if err != nil {
		return e.New(e.BadRequest, err)
	}
	return nil
}

//giteeRequest
//param path : gitea api路径
//param method 请求方式
func giteaRequest(path, method, token string, requestBody []byte) (*http.Response, []byte, error) {
	vcsToken, err := GetVcsToken(token)
	if err != nil {
		return nil, nil, err
	}
	request, er := http.NewRequest(method, path, bytes.NewBuffer(requestBody))
	if er != nil {
		return nil, nil, er
	}
	client := &http.Client{}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", fmt.Sprintf("token %s", vcsToken))
	//request.Body.Read()
	response, err := client.Do(request)

	if err != nil {
		return nil, nil, err
	}

	defer response.Body.Close()

	body, _ := ioutil.ReadAll(response.Body)

	return response, body, nil

}

type giteaUser struct {
	Id int64 `json:"id" form:"id" `
}

func getGiteaUserMe(vcs *models.Vcs) (*giteaUser, error) {
	path := vcs.Address + "/api/v1/user"
	_, body, err := giteaRequest(path, "GET", vcs.VcsToken, nil)
	if err != nil {
		return nil, e.New(e.BadRequest, err)
	}
	user := &giteaUser{}
	_ = json.Unmarshal(body, user)
	return user, nil
}
