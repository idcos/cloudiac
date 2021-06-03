package vcsrv

import (
	"cloudiac/consts"
	"cloudiac/consts/e"
	"cloudiac/models"
	"cloudiac/models/forms"
	"github.com/xanzy/go-gitlab"
	"strings"
)

func ListOrganizationReposById(vcs *models.Vcs,form *forms.GetGitProjectsForm) (projects []*gitlab.Project, total int, err e.Error) {
	git, err := GetGitConn(vcs.VcsToken, vcs.Address)
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
	git, err := GetGitConn(vcs.VcsToken, vcs.Address)
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
	git, err := GetGitConn(vcs.VcsToken, vcs.Address)
	if err != nil {
		return content, err
	}

	opt := &gitlab.GetRawFileOptions{Ref: gitlab.String(form.Branch)}
	row, _, errs := git.RepositoryFiles.GetRawFile(form.RepoId, "README.md", opt)
	if errs != nil {
		return content, e.New(e.GitLabError, err)
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

func TemplateTfvarsSearch(vcs *models.Vcs,repoId uint, repoBranch string) (interface{}, e.Error) {
	tfVarsList :=  make([]string,0)
	var errs error
	if vcs.VcsType == consts.GitLab{
		git, err := GetGitConn(vcs.VcsToken, vcs.Address)
		if err != nil {
			return nil, err
		}
		tfVarsList, errs = getTfvarsList(git, repoBranch, "", repoId)

	}

	if vcs.VcsType == consts.GitEA {
		tfVarsList ,errs = GetGiteaTemplateTfvarsSearch(vcs,repoId,repoBranch,"")
	}

	if errs != nil {
		return nil, e.New(e.GitLabError, errs)
	}

	//c, _, b1 := git.RepositoryFiles.GetFile(564, "state.tf",&sss)
	return tfVarsList, nil
}

func getTfvarsList(git *gitlab.Client, repoBranch, path string, repoId uint) ([]string, error) {
	var fileBlob = "blob"
	var fileTree = "tree"
	pathList := make([]string, 0)
	lto := &gitlab.ListTreeOptions{
		ListOptions: gitlab.ListOptions{Page: 1, PerPage: 1000},
		Ref:         gitlab.String(repoBranch),
		Path:        gitlab.String(path),
	}
	treeNode, _, err := git.Repositories.ListTree(int(repoId), lto)
	if err != nil {
		return nil, err
	}

	for _, i := range treeNode {
		if i.Type == fileBlob && strings.Contains(i.Name, "tfvars") {
			pathList = append(pathList, i.Path)
		}
		if i.Type == fileTree {
			pl, err := getTfvarsList(git, repoBranch, i.Path, repoId)
			if err != nil {
				return nil, err
			}
			pathList = append(pathList, pl...)
		}
	}
	return pathList, nil

}
