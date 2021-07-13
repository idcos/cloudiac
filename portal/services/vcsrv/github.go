package vcsrv

import (
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
	"time"
)

//newGithubInstance
//github api文档: https://docs.github.com/cn/rest/reference/repos
func newGithubInstance(vcs *models.Vcs) (VcsIface, error) {
	return &githubVcs{githubRequest: githubRequest, vcs: vcs}, nil

}

type githubVcs struct {
	githubRequest func(path, method, token string) (*http.Response, []byte, error)
	vcs           *models.Vcs
}

func (github *githubVcs) GetRepo(idOrPath string) (RepoIface, error) {
	path := utils.GenQueryURL(github.vcs.Address, fmt.Sprintf("/repos/%s", idOrPath), nil)
	_, body, er := github.githubRequest(path, "GET", github.vcs.VcsToken)
	if er != nil {
		return nil, e.New(e.BadRequest, er)
	}

	rep := RepositoryGithub{}
	_ = json.Unmarshal(body, &rep)

	return &githubRepoIface{
		githubRequest: github.githubRequest,
		vcs:           github.vcs,
		repository:    &rep,
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
	response, body, err := github.githubRequest(path, "GET", github.vcs.VcsToken)
	if err != nil {
		return nil, 0, e.New(e.BadRequest, err)
	}

	var total int64
	if len(response.Header["X-Total-Count"]) != 0 {
		total, _ = strconv.ParseInt(response.Header["X-Total-Count"][0], 10, 64)
	}
	rep := make([]*RepositoryGithub, 0)
	_ = json.Unmarshal(body, &rep)
	repoList := make([]RepoIface, 0)
	for _, v := range rep {
		repoList = append(repoList, &githubRepoIface{
			githubRequest: github.githubRequest,
			vcs:           github.vcs,
			repository:    v,
			total:         int(total),
		})
	}

	return repoList, total, nil
}

type githubRepoIface struct {
	githubRequest func(path, method, token string) (*http.Response, []byte, error)
	vcs           *models.Vcs
	repository    *RepositoryGithub
	total         int
}

type githubBranch struct {
	Name string `json:"name" form:"name" `
}

func (github *githubRepoIface) ListBranches() ([]string, error) {

	path := utils.GenQueryURL(github.vcs.Address,
		fmt.Sprintf("/repos/%s/branches", github.repository.FullName), nil)
	_, body, err := github.githubRequest(path, "GET", github.vcs.VcsToken)
	if err != nil {
		return nil, e.New(e.BadRequest, err)
	}
	rep := make([]githubBranch, 0)

	_ = json.Unmarshal(body, &rep)
	branchList := []string{}
	for _, v := range rep {
		branchList = append(branchList, v.Name)
	}
	return branchList, nil
}

type githubTag struct {
	Name string `json:"name" form:"name" `
}

func (github *githubRepoIface) ListTags() ([]string, error) {
	path := utils.GenQueryURL(github.vcs.Address, fmt.Sprintf("/repos/%s/%s/tags", github.repository.FullName, github.repository.Name), nil)
	_, body, err := github.githubRequest(path, "GET", github.vcs.VcsToken)
	if err != nil {
		return nil, e.New(e.BadRequest, err)
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
	Commit struct {
		Id string `json:"id" form:"id" `
	} `json:"commit" form:"commit" `
}

func (github *githubRepoIface) BranchCommitId(branch string) (string, error) {
	path := utils.GenQueryURL(github.vcs.Address,
		fmt.Sprintf("/repos/%s/commits/%s", github.repository.FullName, branch), nil)
	_, body, err := github.githubRequest(path, "GET", github.vcs.VcsToken)
	if err != nil {
		return "", e.New(e.BadRequest, err)
	}

	rep := githubCommit{}
	_ = json.Unmarshal(body, &rep)
	return rep.Commit.Id, nil
}

type githubFiles struct {
	Type string `json:"type" form:"type" `
	Path string `json:"path" form:"path" `
	Name string `json:"name" form:"name" `
}

func (github *githubRepoIface) ListFiles(option VcsIfaceOptions) ([]string, error) {
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
	_, body, er := github.githubRequest(path, "GET", github.vcs.VcsToken)
	if er != nil {
		return []string{}, e.New(e.BadRequest, er)
	}
	resp := make([]string, 0)
	rep := make([]githubFiles, 0)
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

func (github *githubRepoIface) ReadFileContent(branch, path string) (content []byte, err error) {
	urlParam := url.Values{}
	urlParam.Set("ref", branch)
	pathAddr := utils.GenQueryURL(github.vcs.Address,
		fmt.Sprintf("/repos/%s/contents/%s", github.repository.FullName, path), urlParam)
	_, body, er := github.githubRequest(pathAddr, "GET", github.vcs.VcsToken)
	if er != nil {
		return nil, e.New(e.BadRequest, er)
	}
	grc := githubReadContent{}
	_ = json.Unmarshal(body[:], &grc)
	decoded, err := base64.StdEncoding.DecodeString(grc.Content)
	if err != nil {
		return nil, e.New(e.BadRequest, er)
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
	}, nil
}

func (github *githubRepoIface) DefaultBranch() string {
	return github.repository.DefaultBranch
}

//giteaRequest
//param path : gitea api路径
//param method 请求方式
func githubRequest(path, method, token string) (*http.Response, []byte, error) {
	request, er := http.NewRequest(method, path, nil)
	if er != nil {
		return nil, nil, er
	}
	client := &http.Client{}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", fmt.Sprintf("token %s", token))
	response, err := client.Do(request)
	if err != nil {
		return nil, nil, err
	}
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	return response, body, nil

}
