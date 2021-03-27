package services

import (
	"cloudiac/configs"
	"cloudiac/consts/e"
	"cloudiac/libs/db"
	"cloudiac/models"
	"cloudiac/models/forms"
	"encoding/json"
	"github.com/xanzy/go-gitlab"
)


func ListOrganizationReposById(tx *db.Session, orgId uint, form *forms.GetGitProjectsForm) (projects []*gitlab.Project, err e.Error) {
	git, err := GetGitConn(tx, orgId)
	if err != nil {
		return nil, err
	}

	opt := &gitlab.ListProjectsOptions{}
	if form.Q != "" {
		opt.Search = gitlab.String(form.Q)
	}

	if form.PageSize_ != 0 && form.CurrentPage_ != 0 {
		opt.Page = form.CurrentPage_
		opt.PerPage = form.PageSize_
	}

	projects, _, er := git.Projects.ListProjects(opt)
	if er != nil {
		return nil, e.New(e.GitLabError, er)
	}

	return
}

func ListRepositoryBranches(tx *db.Session, orgId uint, repoId int) (branches []*gitlab.Branch, err e.Error) {
	git, err := GetGitConn(tx, orgId)
	if err != nil {
		return nil, err
	}

	opt := &gitlab.ListBranchesOptions{}
	branches, _, er := git.Branches.ListBranches(repoId, opt)
	if er != nil {
		return nil, e.New(e.GitLabError, er)
	}
	return branches, nil
}

func GetReadmeContent(tx *db.Session, orgId uint, form *forms.GetReadmeForm) (content models.FileContent, err error) {
	content = models.FileContent{
		Content: "",
	}

	git, err := GetGitConn(tx, orgId)
	if err != nil {
		return content, err
	}

	opt := &gitlab.GetRawFileOptions{Ref: gitlab.String(form.Branch)}
	row, _, err := git.RepositoryFiles.GetRawFile(form.RepoId, "README.md", opt)
	if err != nil {
		return content, err
	}

	res := models.FileContent{
		Content: string(row[:]),
	}
	return res, nil
}

func GetGitConn(tx *db.Session, orgId uint) (git *gitlab.Client, err e.Error) {
	// 优先使用配置文件中的gitlab配置
	conf := configs.Get()
	gitlabUrl := conf.Gitlab.Url
	gitlabToken := conf.Gitlab.Token
	if gitlabUrl == "" || gitlabToken == "" {
		org := models.Organization{}
		if er := tx.Where("id = ?", orgId).First(&org); er != nil {
			return nil, e.New(e.DBError, er)
		}
		var dat map[string]string
		vcsAuthInfo := org.VcsAuthInfo
		if er := json.Unmarshal([]byte(vcsAuthInfo), &dat); er != nil {
			return nil, e.New(e.JSONParseError, er)
		}
		gitlabUrl = dat["url"]
		gitlabToken = dat["token"]
	}

	git, er := gitlab.NewClient(gitlabToken, gitlab.WithBaseURL(gitlabUrl+"/api/v4"))
	if er != nil {
		return nil, e.New(e.JSONParseError, er)
	}
	return
}