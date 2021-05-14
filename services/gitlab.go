package services

import (
	"cloudiac/consts/e"
	"cloudiac/models"
	"cloudiac/models/forms"
	"github.com/xanzy/go-gitlab"
)

func ListOrganizationReposById(vcs *models.Vcs,form *forms.GetGitProjectsForm) (projects []*gitlab.Project, total int, err e.Error) {
	git, err := GetGitConn(vcs.Address, vcs.VcsToken)
	if err != nil {
		return nil, total, err
	}

	opt := &gitlab.ListProjectsOptions{}
	if form.Q != "" {
		opt.Search = gitlab.String(form.Q)
	}

	if form.PageSize_ != 0 && form.CurrentPage_ != 0 {
		opt.Page = form.CurrentPage_
		opt.PerPage = form.PageSize_
	}

	projects, response, er := git.Projects.ListProjects(opt)
	if er != nil {
		return nil, total, e.New(e.GitLabError, er)
	}
	total = response.TotalItems
	return
}

func ListRepositoryBranches(vcs *models.Vcs, form *forms.GetGitBranchesForm) (branches []*gitlab.Branch, err e.Error) {
	git, err := GetGitConn(vcs.Address, vcs.VcsToken)
	if err != nil {
		return nil, err
	}
	opt := &gitlab.ListBranchesOptions{}
	branches, _, er := git.Branches.ListBranches(form.RepoId, opt)
	if er != nil {
		return nil, e.New(e.GitLabError, er)
	}
	return branches, nil
}

func GetReadmeContent(vcs *models.Vcs, form *forms.GetReadmeForm) (content models.FileContent, err error) {
	content = models.FileContent{
		Content: "",
	}

	git, err := GetGitConn(vcs.Address, vcs.VcsToken)
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

func GetGitConn(gitlabToken, gitlabUrl string) (git *gitlab.Client, err e.Error) {
	git, er := gitlab.NewClient(gitlabToken, gitlab.WithBaseURL(gitlabUrl+"/api/v4"))
	if er != nil {
		return nil, e.New(e.JSONParseError, er)
	}
	return
}
