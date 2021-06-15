package vcsrv

import (
	"cloudiac/consts/e"
	"cloudiac/models"
	"cloudiac/utils"
	"cloudiac/utils/logs"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	fPath "path"
	"strconv"
	"strings"
	"time"
)

func newGiteaInstance(vcs *models.Vcs) (VcsIface, error) {
	return &giteaVcs{giteaRequest: giteaRequest, vcs: vcs}, nil

}

type giteaVcs struct {
	giteaRequest func(path, method, token string) (*http.Response, []byte, error)
	vcs          *models.Vcs
}

func (gitea *giteaVcs) GetRepo(idOrPath string) (RepoIface, error) {
	vcsRawPath := GetGiteaUrl(gitea.vcs.Address)
	path := vcsRawPath + fmt.Sprintf("/api/v1/repositories/%s",idOrPath)
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
func (gitea *giteaVcs) ListRepos(namespace, search string, limit, offset int) ([]RepoIface,int64, error) {
	link, _ := url.Parse("/repos/search")
	page := utils.LimitOffset2Page(limit, offset)
	link.RawQuery = fmt.Sprintf("page=%d&limit=%d", page, limit)
	if search != "" {
		link.RawQuery = link.RawQuery + fmt.Sprintf("&q=%s", search)
	}
	path := gitea.vcs.Address + "/api/v1" + link.String()
	response, body, err := gitea.giteaRequest(path, "GET", gitea.vcs.VcsToken)

	if err != nil {
		return nil,0, e.New(e.BadRequest, err)
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

	return repoList,total, nil
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
	var path string
	vcsRawPath := GetGiteaUrl(gitea.vcs.Address)
	if option.Path != "" {
		path = vcsRawPath + "/api/v1" +
			fmt.Sprintf("/repos/%s/contents/%s?limit=0&page=0&ref=%s", gitea.repository.FullName, option.Path, option.Ref)
	} else {
		path = vcsRawPath + "/api/v1" +
			fmt.Sprintf("/repos/%s/contents?limit=0&page=0&ref=%s", gitea.repository.FullName, option.Ref)
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

		matched, err := fPath.Match(option.Search, v.Name)
		if err != nil {
			logs.Get().Debug("file name match err: %v", err)
		}

		if v.Type == "file" && matched {
			resp = append(resp, v.Name)
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

//giteaRequest
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

func DoGiteaRequest(request *http.Request, token string) (*http.Response, error) {
	client := &http.Client{}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", fmt.Sprintf("token %s", token))
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	return response, nil
}
func GetGiteaUrl(address string) string {
	return strings.TrimSuffix(address, "/")
}

func GetGiteaTemplateTfvarsSearch(vcs *models.Vcs, repoId string, repoBranch, filePath string, fileName []string) ([]string, error) {
	repo, err := GetGiteaRepoById(vcs, repoId)
	if err != nil {
		return nil, err
	}
	var path string
	vcsRawPath := GetGiteaUrl(vcs.Address)
	if filePath != "" {
		path = vcsRawPath + "/api/v1" + fmt.Sprintf("/repos/%s/contents/%s?limit=0&page=0", repo, filePath)
	} else {
		path = vcsRawPath + "/api/v1" + fmt.Sprintf("/repos/%s/contents?limit=0&page=0", repo)
	}
	request, er := http.NewRequest("GET", path, nil)
	if er != nil {
		return nil, e.New(e.BadRequest, er)
	}
	response, er := DoGiteaRequest(request, vcs.VcsToken)
	if er != nil {
		return nil, e.New(e.BadRequest, er)
	}
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	resp := make([]string, 0)
	rep := []map[string]interface{}{}
	_ = json.Unmarshal(body, &rep)
	for _, v := range rep {
		if _, ok := v["type"].(string); ok && v["type"].(string) == "dir" {
			repList, _ := GetGiteaTemplateTfvarsSearch(vcs, repoId, repoBranch, v["path"].(string), fileName)
			resp = append(resp, repList...)
		}

		if _, ok := v["type"].(string); ok && v["type"].(string) == "file" &&
			utils.ArrayIsHasSuffix(fileName, v["name"].(string)) {
			resp = append(resp, v["name"].(string))
		}

	}

	return resp, nil

}

func GetGiteaRepoById(vcs *models.Vcs, repoId string) (string, e.Error) {
	vcsRawPath := GetGiteaUrl(vcs.Address)
	path := vcsRawPath + fmt.Sprintf("/api/v1/repositories/%s", repoId)
	request, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return "", e.New(e.BadRequest, err)
	}
	response, err := DoGiteaRequest(request, vcs.VcsToken)
	if response == nil || err != nil {
		return "", e.New(e.BadRequest, err)
	}
	defer response.Body.Close()
	repo := map[string]interface{}{}
	body, _ := ioutil.ReadAll(response.Body)
	json.Unmarshal(body, &repo)

	if d, ok := repo["full_name"].(string); ok {
		return d, nil
	}
	return "", nil

}

func GetGiteaBranchCommitId(vcs *models.Vcs, repoId string, repoBranch string) (string, error) {
	repo, err := GetGiteaRepoById(vcs, repoId)
	if err != nil {
		return "", err
	}
	vcsRawPath := GetGiteaUrl(vcs.Address)
	path := vcsRawPath + "/api/v1" + fmt.Sprintf("/repos/%s/branches/%s?limit=0&page=0", repo, repoBranch)
	request, er := http.NewRequest("GET", path, nil)
	if er != nil {
		return "", e.New(e.BadRequest, er)
	}
	response, er := DoGiteaRequest(request, vcs.VcsToken)
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	rep := map[string]interface{}{}
	json.Unmarshal(body, &rep)
	//return branchList, nil
	var commit string
	if _, ok := rep["commit"].(map[string]interface{}); ok {
		commit = rep["commit"].(map[string]interface{})["id"].(string)
	}
	return commit, nil

}
