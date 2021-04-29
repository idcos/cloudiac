package apps

import (
	"cloudiac/configs"
	"cloudiac/consts/e"
	"cloudiac/libs/page"
	"cloudiac/models"
	"cloudiac/models/forms"
	"cloudiac/services"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

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
}

func GetGiteaReadme(form *forms.GetReadmeForm) (interface{}, e.Error) {
	conf := configs.Get()
	gitlabUrl := conf.Gitea.Url
	repo, err := GetGiteaRepoById(form.RepoId)
	if err != nil {
		return nil, err
	}
	path := gitlabUrl + "/api/v1" + fmt.Sprintf("repos/%s/raw/README.md?ref=%s", repo, form.Branch)
	request, er := http.NewRequest("GET", path, nil)
	if er != nil {
		return nil, e.New(e.BadRequest, err)
	}
	response, er := services.DoGiteaRequest(request, conf.Gitea.Token)
	body, _ := ioutil.ReadAll(response.Body)

	res := models.FileContent{
		Content: string(body[:]),
	}
	return res, nil
}

func ListGiteaRepoBranches(form *forms.GetGitBranchesForm) ([]*Branches, e.Error) {
	conf := configs.Get()
	gitlabUrl := conf.Gitea.Url
	repo, err := GetGiteaRepoById(form.RepoId)
	if err != nil {
		return nil, err
	}
	path := gitlabUrl + "/api/v1" + fmt.Sprintf("/repos/%s/branches?limit=0&page=0\n", repo)
	request, er := http.NewRequest("GET", path, nil)
	if err != nil {
		return nil, e.New(e.BadRequest, er)
	}
	response, er := services.DoGiteaRequest(request, conf.Gitea.Token)
	body, _ := ioutil.ReadAll(response.Body)
	rep := []map[string]interface{}{}
	json.Unmarshal(body, &rep)
	branchList := []*Branches{}
	for _, v := range rep {
		branch := &Branches{Name: v["name"].(string)}
		branchList = append(branchList, branch)
	}
	return branchList, nil

}

func GetGiteaRepoById(repoId int) (string, e.Error) {
	conf := configs.Get()
	path := fmt.Sprintf("http://10.0.2.135:10080/api/v1/repositories/%d", repoId)
	request, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return "", e.New(e.BadRequest, err)
	}
	response, err := services.DoGiteaRequest(request, conf.Gitea.Token)
	repo := map[string]interface{}{}
	body, _ := ioutil.ReadAll(response.Body)
	json.Unmarshal(body, &repo)
	return repo["full_name"].(string), nil

}

func ListGiteaOrganizationRepos(form *forms.GetGitProjectsForm) (interface{}, e.Error) {
	conf := configs.Get()
	gitlabUrl := conf.Gitea.Url
	link, _ := url.Parse("/repos/search")
	link.RawQuery = fmt.Sprintf("page=%d&limit=%d", form.CurrentPage(), form.PageSize())
	path := gitlabUrl + "/api/v1" + link.String()
	request, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return nil, e.New(e.BadRequest, err)
	}
	response, err := services.DoGiteaRequest(request, conf.Gitea.Token)
	var total int64
	if len(response.Header["X-Total-Count"]) != 0 {
		total, _ = strconv.ParseInt(response.Header["X-Total-Count"][0], 10, 64)
	}
	if err != nil {
		return nil, e.New(e.BadRequest, err)
	}
	body, _ := ioutil.ReadAll(response.Body)
	rep := SearchRepoResponse{}
	json.Unmarshal(body, &rep)
	projectList := []*Projects{}
	for _, v := range rep.Repos {
		project := &Projects{
			ID:             int(v.ID),
			Description:    v.Description,
			DefaultBranch:  v.DefaultBranch,
			SSHURLToRepo:   v.SSHURL,
			HTTPURLToRepo:  v.CloneURL,
			Name:           v.Name,
			LastActivityAt: &v.Updated,
		}
		projectList = append(projectList, project)
	}

	return page.PageResp{
		Total:    total,
		PageSize: form.CurrentPage(),
		List:     projectList,
	}, nil

}
