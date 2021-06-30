package vcsrv

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/models"
	"cloudiac/utils"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

func newGiteaInstance(vcs *models.Vcs) (VcsIface, error) {
	vcs.Address = utils.GetUrl(vcs.Address)
	return &giteaVcs{giteaRequest: giteaRequest, vcs: vcs}, nil

}

type giteaVcs struct {
	giteaRequest func(path, method, token string) (*http.Response, []byte, error)
	vcs          *models.Vcs
}

func (gitea *giteaVcs) GetRepo(idOrPath string) (RepoIface, error) {
	path := gitea.vcs.Address + fmt.Sprintf("/api/v1/repositories/%s", idOrPath)
	response, body, er := gitea.giteaRequest(path, "GET", gitea.vcs.VcsToken)
	if er != nil {
		return nil, e.New(e.BadRequest, er)
	}
	defer response.Body.Close()
	rep := Repository{}
	_ = json.Unmarshal(body, &rep)
	return &giteaRepoIface{
		giteaRequest: gitea.giteaRequest,
		vcs:          gitea.vcs,
		repository:   &rep,
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

//Fixme ListRepos中的数据不能直接调用repo接口的方法
func (gitea *giteaVcs) ListRepos(namespace, search string, limit, offset int) ([]RepoIface, int64, error) {
	link, _ := url.Parse("/repos/search")
	page := utils.LimitOffset2Page(limit, offset)
	link.RawQuery = fmt.Sprintf("page=%d&limit=%d", page, limit)
	if search != "" {
		link.RawQuery = link.RawQuery + fmt.Sprintf("&q=%s", search)
	}
	path := gitea.vcs.Address + "/api/v1" + link.String()
	response, body, err := gitea.giteaRequest(path, "GET", gitea.vcs.VcsToken)

	if err != nil {
		return nil, 0, e.New(e.BadRequest, err)
	}

	defer response.Body.Close()
	var total int64
	if len(response.Header["X-Total-Count"]) != 0 {
		total, _ = strconv.ParseInt(response.Header["X-Total-Count"][0], 10, 64)
	}

	rep := SearchRepoResponse{}
	_ = json.Unmarshal(body, &rep)

	repoList := make([]RepoIface, 0)
	for _, v := range rep.Repos {
		repoList = append(repoList, &giteaRepoIface{
			giteaRequest: gitea.giteaRequest,
			vcs:          gitea.vcs,
			repository:   v,
		})
	}

	return repoList, total, nil
}

type giteaRepoIface struct {
	giteaRequest func(path, method, token string) (*http.Response, []byte, error)
	vcs          *models.Vcs
	repository   *Repository
}

type giteaBranch struct {
	Name string `json:"name" form:"name" `
}

func (gitea *giteaRepoIface) ListBranches() ([]string, error) {
	path := gitea.vcs.Address + "/api/v1" +
		fmt.Sprintf("/repos/%s/branches?limit=0&page=0", gitea.repository.FullName)

	response, body, err := gitea.giteaRequest(path, "GET", gitea.vcs.VcsToken)
	if err != nil {
		return nil, e.New(e.BadRequest, err)
	}
	defer response.Body.Close()
	rep := make([]giteaBranch, 0)

	_ = json.Unmarshal(body, &rep)
	branchList := []string{}
	for _, v := range rep {
		branchList = append(branchList, v.Name)
	}
	return branchList, nil
}

type giteaCommit struct {
	Commit struct {
		Id string `json:"id" form:"id" `
	} `json:"commit" form:"commit" `
}

func (gitea *giteaRepoIface) BranchCommitId(branch string) (string, error) {
	path := gitea.vcs.Address + "/api/v1" +
		fmt.Sprintf("/repos/%s/branches/%s?limit=0&page=0", gitea.repository.FullName, branch)
	response, body, err := gitea.giteaRequest(path, "GET", gitea.vcs.VcsToken)
	if err != nil {
		return "", e.New(e.BadRequest, err)
	}
	defer response.Body.Close()
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
		path += "/api/v1" +
			fmt.Sprintf("/repos/%s/contents/%s?limit=0&page=0&ref=%s",
				gitea.repository.FullName, option.Path, branch)
	} else {
		path += "/api/v1" +
			fmt.Sprintf("/repos/%s/contents?limit=0&page=0&ref=%s",
				gitea.repository.FullName, branch)
	}
	response, body, er := gitea.giteaRequest(path, "GET", gitea.vcs.VcsToken)
	if er != nil {
		return []string{}, e.New(e.BadRequest, er)
	}
	defer response.Body.Close()
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
	pathAddr := gitea.vcs.Address + "/api/v1" +
		fmt.Sprintf("/repos/%s/raw/%s?ref=%s", gitea.repository.FullName, path, branch)
	response, body, er := gitea.giteaRequest(pathAddr, "GET", gitea.vcs.VcsToken)
	if er != nil {
		return []byte{}, e.New(e.BadRequest, er)
	}
	defer response.Body.Close()

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
	}, nil
}

func (gitea *giteaRepoIface) DefaultBranch() string {
	return gitea.repository.DefaultBranch
}

//giteeRequest
//param path : gitea api路径
//param method 请求方式
func giteaRequest(path, method, token string) (*http.Response, []byte, error) {
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
	body, _ := ioutil.ReadAll(response.Body)
	return response, body, nil

}
