package services

import (
	"cloudiac/consts/e"
	"cloudiac/models"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

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

func GetGiteaTemplateTfvarsSearch(vcs *models.Vcs, repoId uint, repoBranch,filePath,fileName string) ([]string, error) {
	repo, err := GetGiteaRepoById(vcs, int(repoId))
	if err != nil {
		return nil, err
	}
	var path string
	if filePath!="" {
		path = vcs.Address + "/api/v1" + fmt.Sprintf("/repos/%s/contents/%s?limit=0&page=0", repo,filePath)
	}else {
		path = vcs.Address + "/api/v1" + fmt.Sprintf("/repos/%s/contents?limit=0&page=0", repo)
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
			repList, _ := GetGiteaTemplateTfvarsSearch(vcs, repoId, repoBranch,v["path"].(string),fileName)
			resp = append(resp, repList...)
		}

		if _, ok := v["type"].(string); ok && v["type"].(string) == "file" && strings.Contains(v["name"].(string), fileName) {
			resp = append(resp, v["name"].(string))
		}

	}

	return resp, nil

}

func GetGiteaRepoById(vcs *models.Vcs, repoId int) (string, e.Error) {

	path := vcs.Address + fmt.Sprintf("/api/v1/repositories/%d", repoId)
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

func GetGiteaBranchCommitId(vcs *models.Vcs, repoId uint, repoBranch string) (string, error) {
	repo, err := GetGiteaRepoById(vcs, int(repoId))
	if err != nil {
		return "", err
	}
	path := vcs.Address + "/api/v1" + fmt.Sprintf("/repos/%s/branches/%s?limit=0&page=0", repo, repoBranch)
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
	if _, ok := rep["commit"].(string); ok {
		commit = rep["commit"].(string)
	}
	return commit, nil

}
