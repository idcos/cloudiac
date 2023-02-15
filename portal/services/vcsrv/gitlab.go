// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package vcsrv

import (
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/models"
	"cloudiac/utils"
	"fmt"
	"net/url"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/xanzy/go-gitlab"
)

func newGitlabInstance(vcs *models.Vcs) (VcsIface, error) {
	gitConn, err := GetGitConn(vcs.VcsToken, vcs.Address)
	if err != nil {
		return nil, err
	}
	return &gitlabVcsIface{
		vcs:     vcs,
		gitConn: gitConn,
	}, nil
}

type gitlabVcsIface struct {
	vcs     *models.Vcs
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

func (git *gitlabVcsIface) UserInfo() (UserInfo, error) {

	return UserInfo{}, nil
}

func (git *gitlabVcsIface) TokenCheck() error {
	opt := &gitlab.ListProjectsOptions{}
	opt.Page = utils.LimitOffset2Page(1, 1)

	_, response, err := git.gitConn.Projects.ListProjects(opt)
	if err != nil {
		if strings.Contains(err.Error(), "Unauthorized") {
			return e.New(e.VcsInvalidToken, err)
		}
		return err
	}

	if response.StatusCode > 300 {
		return e.New(e.VcsInvalidToken, fmt.Sprintf("token valid check response code: %d", response.StatusCode))
	}

	return nil
}

func (v *gitlabVcsIface) RepoBaseHttpAddr() string {
	return v.vcs.Address
}

type gitlabRepoIface struct {
	gitConn *gitlab.Client
	Project *gitlab.Project
}

func (git *gitlabRepoIface) ListBranches() ([]string, error) {
	branchList := make([]string, 0)
	currentPage := 1

	for {
		branches, response, er := git.gitConn.Branches.ListBranches(git.Project.ID,
			&gitlab.ListBranchesOptions{
				ListOptions: gitlab.ListOptions{
					Page:    currentPage,
					PerPage: consts.GitlabPerPageSize,
				},
			})
		if er != nil {
			return nil, e.New(e.VcsError, er)
		}

		for _, branch := range branches {
			branchList = append(branchList, branch.Name)
		}

		if currentPage == response.TotalPages {
			break
		}
		currentPage++
	}

	return branchList, nil
}

func (git *gitlabRepoIface) ListTags() ([]string, error) {
	tagList := make([]string, 0)
	currentPage := 1

	for {
		tags, response, er := git.gitConn.Tags.ListTags(git.Project.ID,
			&gitlab.ListTagsOptions{ListOptions: gitlab.ListOptions{
				Page:    currentPage,
				PerPage: consts.GitlabPerPageSize,
			}})
		if er != nil {
			return nil, e.New(e.VcsError, er)
		}

		for _, tag := range tags {
			tagList = append(tagList, tag.Name)
		}

		if currentPage == response.TotalPages {
			break
		}
		currentPage++
	}

	return tagList, nil
}

func (git *gitlabRepoIface) BranchCommitId(branch string) (string, error) {
	lco := &gitlab.ListCommitsOptions{
		RefName: gitlab.String(branch),
	}
	commits, _, commitErr := git.gitConn.Commits.ListCommits(git.Project.ID, lco)
	if commitErr != nil {
		return "nil", e.New(e.VcsError, commitErr)
	}
	if len(commits) > 0 {
		return commits[0].ID, nil
	}
	return "", e.New(e.VcsError, fmt.Errorf("repo %s, commit is null", git.Project.Name))
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
	pathList, err := git.UpdatePlaybookWorkDir(pathList, []byte(getBranch(git, option.Ref)), option, consts.PlaybookDir)
	if err != nil {
		return nil, err
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

func (git *gitlabRepoIface) UpdatePlaybookWorkDir(resp []string, body []byte, option VcsIfaceOptions, pattern string) ([]string, error) {
	if strings.Contains(option.Path, pattern) {
		branch := strings.TrimSpace(string(body))
		row, err := git.ReadFileContent(branch, option.Path)
		if err != nil {
			return resp, nil
		}
		workDir := strings.TrimSpace(string(row))
		option.Path = strings.Replace(option.Path, pattern, "", 1)
		option.Path = filepath.Join(option.Path, workDir)
		pl, err := git.ListFiles(option)
		if err != nil {
			return resp, nil
		}
		resp = append(resp, pl...)
	}
	return resp, nil
}

func (git *gitlabRepoIface) JudgeFileType(branch, workdir, filename string) (pathname string, err error) {
	lto := &gitlab.ListTreeOptions{
		ListOptions: gitlab.ListOptions{Page: 1, PerPage: 1000},
		Ref:         gitlab.String(branch),
		Path:        gitlab.String(workdir),
	}
	treeNode, _, err := git.gitConn.Repositories.ListTree(git.Project.ID, lto)
	if err != nil {
		return filename, err
	}
	for _, i := range treeNode {
		if i.Mode == "120000" {
			content, err := git.ReadFileContent(branch, i.Path)
			if err != nil {
				return filename, nil
			}
			if i.Name == consts.PlaybookDir {
				count := strings.Count(strings.TrimSpace(string(content)), "../")
				filename = fmt.Sprintf("%s%s", strings.Repeat("../", count), filename)
				return filename, nil
			}
			filename = strings.TrimSpace(string(content))
			return filename, nil
		}
	}
	return filename, nil
}

func (git *gitlabRepoIface) ReadFileContent(branch, path string) ([]byte, error) {
	opt := &gitlab.GetRawFileOptions{Ref: gitlab.String(branch)}
	row, _, err := git.gitConn.RepositoryFiles.GetRawFile(git.Project.ID, path, opt)
	if err != nil && strings.Contains(err.Error(), "File Not Found") {
		return nil, e.New(e.ObjectNotExists, err)
	}
	return row, err
}

type Projects struct {
	ID             string     `json:"id"`
	Description    string     `json:"description"`
	DefaultBranch  string     `json:"default_branch"`
	SSHURLToRepo   string     `json:"ssh_url_to_repo"`
	HTTPURLToRepo  string     `json:"http_url_to_repo"`
	Name           string     `json:"name"`
	LastActivityAt *time.Time `json:"last_activity_at,omitempty"`
	FullName       string     `json:"fullName"`
}

func (git *gitlabRepoIface) FormatRepoSearch() (project *Projects, err e.Error) {
	return &Projects{
		ID:             strconv.Itoa(git.Project.ID),
		Description:    git.Project.Description,
		DefaultBranch:  git.Project.DefaultBranch,
		SSHURLToRepo:   git.Project.SSHURLToRepo,
		HTTPURLToRepo:  git.Project.HTTPURLToRepo,
		Name:           git.Project.Name,
		LastActivityAt: git.Project.LastActivityAt,
		FullName:       git.Project.PathWithNamespace,
	}, nil
}

func (git *gitlabRepoIface) DefaultBranch() string {
	return git.Project.DefaultBranch
}

func (git *gitlabRepoIface) AddWebhook(url string) error {
	_, _, err := git.gitConn.Projects.AddProjectHook(git.Project.ID, &gitlab.AddProjectHookOptions{
		URL:                 gitlab.String(url),
		PushEvents:          gitlab.Bool(true),
		MergeRequestsEvents: gitlab.Bool(true),
	})
	return err
}

type ProjectsHook struct {
	ID                       int        `json:"id"`
	URL                      string     `json:"url"`
	ConfidentialNoteEvents   bool       `json:"confidential_note_events"`
	ProjectID                int        `json:"project_id"`
	PushEvents               bool       `json:"push_events"`
	PushEventsBranchFilter   string     `json:"push_events_branch_filter"`
	IssuesEvents             bool       `json:"issues_events"`
	ConfidentialIssuesEvents bool       `json:"confidential_issues_events"`
	MergeRequestsEvents      bool       `json:"merge_requests_events"`
	TagPushEvents            bool       `json:"tag_push_events"`
	NoteEvents               bool       `json:"note_events"`
	JobEvents                bool       `json:"job_events"`
	PipelineEvents           bool       `json:"pipeline_events"`
	WikiPageEvents           bool       `json:"wiki_page_events"`
	DeploymentEvents         bool       `json:"deployment_events"`
	EnableSSLVerification    bool       `json:"enable_ssl_verification"`
	CreatedAt                *time.Time `json:"created_at"`
}

func (git *gitlabRepoIface) ListWebhook() ([]RepoHook, error) {
	resp := make([]RepoHook, 0)
	projectsHook, _, err := git.gitConn.Projects.ListProjectHooks(git.Project.ID, &gitlab.ListProjectHooksOptions{})
	for _, p := range projectsHook {
		resp = append(resp, RepoHook{
			Id:  p.ID,
			Url: p.URL,
		})
	}
	return resp, err
}

func (git *gitlabRepoIface) DeleteWebhook(id int) error {
	_, err := git.gitConn.Projects.DeleteProjectHook(git.Project.ID, id)
	return err
}

func (git *gitlabRepoIface) CreatePrComment(prId int, comment string) error {
	if _, _, err := git.gitConn.Notes.CreateMergeRequestNote(git.Project.ID, prId, &gitlab.CreateMergeRequestNoteOptions{Body: gitlab.String(comment)}); err != nil {
		return err
	}
	return nil
}

func (git *gitlabRepoIface) GetFullFilePath(address, filePath, repoRevision string) string {
	u, _ := url.Parse(address)
	u.Path = path.Join(u.Path, git.Project.PathWithNamespace, "-/blob", repoRevision, filePath)
	return u.String()
}

func (git *gitlabRepoIface) GetCommitFullPath(address, commitId string) string {
	u, _ := url.Parse(address)
	u.Path = path.Join(u.Path, git.Project.PathWithNamespace, "commit", commitId)
	return u.String()
}

func GetGitConn(gitlabToken, gitlabUrl string) (*gitlab.Client, e.Error) {
	token, err := GetVcsToken(gitlabToken)
	if err != nil {
		return nil, e.New(e.VcsError, err)
	}
	git, er := gitlab.NewClient(token, gitlab.WithBaseURL(gitlabUrl+"/api/v4"))
	if er != nil {
		return nil, e.New(e.JSONParseError, er)
	}
	return git, nil
}
