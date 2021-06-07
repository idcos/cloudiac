package vcsrv

import (
	"cloudiac/consts"
	"cloudiac/consts/e"
	"cloudiac/models"
	"cloudiac/models/forms"
	"cloudiac/utils"
	"encoding/json"
	"fmt"
	"github.com/xanzy/go-gitlab"
	"time"
)

func newGitlabInstance(vcs *models.Vcs) (VcsIface, error) {
	gitConn, err := GetGitConn(vcs.VcsToken, vcs.Address)
	if err != nil {
		return nil, err
	}
	return &gitlabVcsIface{gitConn: gitConn}, nil
}

type gitlabVcsIface struct {
	gitConn *gitlab.Client
}

func (git *gitlabVcsIface) GetRepo(option VcsIfaceOptions) (RepoIface, error) {
	project, response, err := git.gitConn.Projects.GetProject(option.IdOrPath, nil)
	if err != nil {
		return nil, err
	}
	return &gitlabRepoIface{
		gitConn: git.gitConn,
		Project: project,
		Total:   response.TotalItems,
	}, nil
}
func (git *gitlabVcsIface) ListRepos(option VcsIfaceOptions) ([]RepoIface, error) {
	opt := &gitlab.ListProjectsOptions{}

	if option.Search != "" {
		opt.Search = gitlab.String(option.Search)
	}

	if option.Limit != 0 && option.Offset != 0 {
		opt.Page = option.Offset
		opt.PerPage = option.Limit
	}

	projects, response, err := git.gitConn.Projects.ListProjects(opt)
	if err != nil {
		return nil, err
	}

	repoList := make([]RepoIface, 0)
	for _, project := range projects {
		repoList = append(repoList, &gitlabRepoIface{
			gitConn: git.gitConn,
			Project: project,
			Total:   response.TotalItems,
		})
	}

	return repoList, nil
}

type gitlabRepoIface struct {
	gitConn *gitlab.Client
	Project *gitlab.Project
	Total   int
}

func (git *gitlabRepoIface) ListBranches(option VcsIfaceOptions) ([]string, error) {
	branchList := make([]string, 0)
	opt := &gitlab.ListBranchesOptions{}
	branches, _, er := git.gitConn.Branches.ListBranches(git.Project.ID, opt)
	if er != nil {
		return nil, e.New(e.GitLabError, er)
	}
	for _, branch := range branches {
		branchList = append(branchList, branch.Name)
	}
	return branchList, nil
}
func (git *gitlabRepoIface) BranchCommitId(option VcsIfaceOptions) (string, error) {
	commits, _, commitErr := git.gitConn.Commits.ListCommits(git.Project.ID, &gitlab.ListCommitsOptions{})
	if commitErr != nil {
		return "nil", e.New(e.GitLabError, commitErr)
	}
	if commits != nil {
		return commits[0].ID, nil
	}
	return "", e.New(e.GitLabError, fmt.Errorf("repo %s, commit is null", git.Project.Name))
}
func (git *gitlabRepoIface) ListFiles(option VcsIfaceOptions) ([]string, error) {

	var (
		fileBlob = "blob"
		fileTree = "tree"
	)
	pathList := make([]string, 0)
	lto := &gitlab.ListTreeOptions{
		ListOptions: gitlab.ListOptions{Page: 1, PerPage: 1000},
		Ref:         gitlab.String(option.Branch),
		Path:        gitlab.String(option.Path),
	}
	treeNode, _, err := git.gitConn.Repositories.ListTree(git.Project.ID, lto)
	if err != nil {
		return nil, err
	}

	for _, i := range treeNode {
		if i.Type == fileBlob && utils.ArrayIsHasSuffix(option.IsHasSuffixFileName, i.Name)  {
			pathList = append(pathList, i.Path)
		}
		if i.Type == fileTree && option.Recursive {
			option.Path = i.Path
			pl, err := git.ListFiles(option)
			if err != nil {
				return nil, err
			}
			pathList = append(pathList, pl...)
		}
	}
	return pathList, nil

}
func (git *gitlabRepoIface) ReadFileContent(option VcsIfaceOptions) (content []byte, err error) {
	opt := &gitlab.GetRawFileOptions{Ref: gitlab.String(option.Branch)}
	row, _, errs := git.gitConn.RepositoryFiles.GetRawFile(git.Project.ID, option.Path, opt)
	if errs != nil {
		return content, e.New(e.GitLabError, err)
	}
	return row, nil
}

type Projects struct {
	ID             int        `json:"id"`
	Description    string     `json:"description"`
	DefaultBranch  string     `json:"default_branch"`
	SSHURLToRepo   string     `json:"ssh_url_to_repo"`
	HTTPURLToRepo  string     `json:"http_url_to_repo"`
	Name           string     `json:"name"`
	LastActivityAt *time.Time `json:"last_activity_at,omitempty"`
}

func (gitlab *gitlabRepoIface) FormatRepoSearch(option VcsIfaceOptions) (project *Projects, err e.Error) {
	jsonProjects, er := json.Marshal(gitlab.Project)
	if er != nil {
		return nil, e.New(e.JSONParseError, er)
	}
	repos := &Projects{}
	er = json.Unmarshal(jsonProjects, &repos)
	if er != nil {
		return nil, e.New(e.JSONParseError, er)
	}

	return repos, nil
}

func ListOrganizationReposById(vcs *models.Vcs, form *forms.GetGitProjectsForm) (projects []*gitlab.Project, total int, err e.Error) {
	git, err := GetGitConn(vcs.VcsToken, vcs.Address)
	if err != nil {
		return nil, total, err
	}

	opt := &gitlab.ListProjectsOptions{}
	if form.Q != "" {
		opt.Search = gitlab.String(form.Q)
	}

	if form.PageSize_ != 0 && form.CurrentPage_ != 0 {
		opt.Search = gitlab.String(form.Q)
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

func TemplateTfvarsSearch(vcs *models.Vcs, repoId uint, repoBranch string, fileName []string) (interface{}, e.Error) {
	tfVarsList := make([]string, 0)
	var errs error
	if vcs.VcsType == consts.GitLab {
		git, err := GetGitConn(vcs.VcsToken, vcs.Address)
		if err != nil {
			return nil, err
		}
		tfVarsList, errs = getTfvarsList(git, repoBranch, "", repoId, fileName)

	}

	if vcs.VcsType == consts.GitEA {
		tfVarsList, errs = GetGiteaTemplateTfvarsSearch(vcs, repoId, repoBranch, "", fileName)
	}

	if errs != nil {
		return nil, e.New(e.GitLabError, errs)
	}

	//c, _, b1 := git.RepositoryFiles.GetFile(564, "state.tf",&sss)
	return tfVarsList, nil
}

func getTfvarsList(git *gitlab.Client, repoBranch, path string, repoId uint, fileName []string) ([]string, error) {
	var (
		fileBlob = "blob"
		fileTree = "tree"
	)
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
		if i.Type == fileBlob && utils.ArrayIsHasSuffix(fileName, i.Name) {
			pathList = append(pathList, i.Path)
		}
		if i.Type == fileTree {
			pl, err := getTfvarsList(git, repoBranch, i.Path, repoId, fileName)
			if err != nil {
				return nil, err
			}
			pathList = append(pathList, pl...)
		}
	}
	return pathList, nil

}


