package vcsrv

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/models"
	"cloudiac/utils"
	"fmt"
	"github.com/xanzy/go-gitlab"
	"strconv"
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
	project, _, err := git.gitConn.Projects.GetProject(idOrPath, nil)
	if err != nil {
		return nil, err
	}
	return &gitlabRepoIface{
		gitConn: git.gitConn,
		Project: project,
	}, nil
}
func (git *gitlabVcsIface) ListRepos(namespace, search string, limit, offset int) ([]RepoIface, int64, error) {
	opt := &gitlab.ListProjectsOptions{}

	if search != "" {
		opt.Search = gitlab.String(search)
	}

	if limit != 0 && offset != 0 {
		opt.Page = utils.LimitOffset2Page(limit, offset)
		opt.PerPage = limit
	}

	projects, response, err := git.gitConn.Projects.ListProjects(opt)
	if err != nil {
		return nil, 0, err
	}

	repoList := make([]RepoIface, 0)
	for _, project := range projects {
		repoList = append(repoList, &gitlabRepoIface{
			gitConn: git.gitConn,
			Project: project,
		})
	}
	return repoList, int64(response.TotalItems), nil
}

type gitlabRepoIface struct {
	gitConn *gitlab.Client
	Project *gitlab.Project
}

func (git *gitlabRepoIface) ListBranches() ([]string, error) {
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

func (git *gitlabRepoIface) ListTags() ([]string, error) {
	tagList := make([]string, 0)
	opt := &gitlab.ListTagsOptions{}
	tags, _, er := git.gitConn.Tags.ListTags(git.Project.ID, opt)
	if er != nil {
		return nil, e.New(e.GitLabError, er)
	}
	for _, branch := range tags {
		tagList = append(tagList, branch.Name)
	}
	return tagList, nil
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
		Ref:         gitlab.String(getBranch(git, option.Ref)),
		Path:        gitlab.String(option.Path),
	}
	treeNode, _, err := git.gitConn.Repositories.ListTree(git.Project.ID, lto)
	if err != nil {
		return nil, err
	}

	for _, i := range treeNode {
		if i.Type == fileBlob && matchGlob(option.Search, i.Name) {
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
	ID             string     `json:"id"`
	Description    string     `json:"description"`
	DefaultBranch  string     `json:"default_branch"`
	SSHURLToRepo   string     `json:"ssh_url_to_repo"`
	HTTPURLToRepo  string     `json:"http_url_to_repo"`
	Name           string     `json:"name"`
	LastActivityAt *time.Time `json:"last_activity_at,omitempty"`
}

func (gitlab *gitlabRepoIface) FormatRepoSearch() (project *Projects, err e.Error) {
	return &Projects{
		ID:             strconv.Itoa(gitlab.Project.ID),
		Description:    gitlab.Project.Description,
		DefaultBranch:  gitlab.Project.DefaultBranch,
		SSHURLToRepo:   gitlab.Project.SSHURLToRepo,
		HTTPURLToRepo:  gitlab.Project.HTTPURLToRepo,
		Name:           gitlab.Project.Name,
		LastActivityAt: gitlab.Project.LastActivityAt,
	}, nil
}

func (gitlab *gitlabRepoIface) DefaultBranch() string {
	return gitlab.Project.DefaultBranch
}

func GetGitConn(gitlabToken, gitlabUrl string) (git *gitlab.Client, err e.Error) {
	git, er := gitlab.NewClient(gitlabToken, gitlab.WithBaseURL(gitlabUrl+"/api/v4"))
	if er != nil {
		return nil, e.New(e.JSONParseError, er)
	}
	return
}
