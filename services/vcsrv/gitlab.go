package vcsrv

import (
	"cloudiac/consts"
	"cloudiac/consts/e"
	"cloudiac/models"
	"cloudiac/models/forms"
	"cloudiac/utils"
	"cloudiac/utils/logs"
	"encoding/json"
	"fmt"
	"github.com/xanzy/go-gitlab"
	"path"
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

func (git *gitlabVcsIface) GetRepo(idOrPath string) (RepoIface, error) {
	project, response, err := git.gitConn.Projects.GetProject(idOrPath, nil)
	if err != nil {
		return nil, err
	}
	return &gitlabRepoIface{
		gitConn: git.gitConn,
		Project: project,
		Total:   response.TotalItems,
	}, nil
}
func (git *gitlabVcsIface) ListRepos(namespace, search string, limit, offset uint) ([]RepoIface, error) {
	opt := &gitlab.ListProjectsOptions{}

	if search != "" {
		opt.Search = gitlab.String(search)
	}

	if limit != 0 && offset != 0 {
		opt.Page = int(offset)
		opt.PerPage = int(limit)
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

func (git *gitlabRepoIface) ListBranches(search string, limit, offset uint) ([]string, error) {
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
func (git *gitlabRepoIface) BranchCommitId(branch string) (string, error) {
	lco := &gitlab.ListCommitsOptions{
		RefName: gitlab.String(branch),
	}
	commits, _, commitErr := git.gitConn.Commits.ListCommits(git.Project.ID, lco)
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
		matched, err := path.Match(option.Search, i.Name)
		if err != nil {
			logs.Get().Debug("file name match err: %v", err)
		}
		if i.Type == fileBlob && matched {
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
func (git *gitlabRepoIface) ReadFileContent(branch, path string) (content []byte, err error) {
	opt := &gitlab.GetRawFileOptions{Ref: gitlab.String(branch)}
	row, _, errs := git.gitConn.RepositoryFiles.GetRawFile(git.Project.ID, path, opt)
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

func (gitlab *gitlabRepoIface) FormatRepoSearch() (project *Projects, err e.Error) {
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
