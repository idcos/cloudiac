package apps

import (
	"cloudiac/consts"
	"cloudiac/consts/e"
	"cloudiac/libs/ctx"
	"cloudiac/libs/page"
	"cloudiac/models"
	"cloudiac/models/forms"
	"cloudiac/services"
	vcs2 "cloudiac/services/vcsrv"
	"cloudiac/utils"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/xanzy/go-gitlab"
	"path/filepath"
	"time"
)

type SearchTaskResp struct {
	models.Task
	RepoBranch  string `json:"repoBranch" form:"repoBranch" `
	CreatorName string `json:"creatorName" form:"creatorName" `
	CreatedTime int64  `json:"createdTime" form:"createdTime" `
	EndTime     int64  `json:"endTime" form:"endTime" `
}

func SearchTask(c *ctx.ServiceCtx, form *forms.SearchTaskForm) (interface{}, e.Error) {
	tx := c.DB().Debug()
	query := services.QueryTask(tx, form.Status, form.Q, form.TemplateId)
	p := page.New(form.CurrentPage(), form.PageSize(), query)
	taskResp := make([]*SearchTaskResp, 0)
	if err := p.Scan(&taskResp); err != nil {
		return nil, e.New(e.DBError, err)
	}

	for _, resp := range taskResp {
		user, err := services.GetUserById(tx, resp.Creator)
		if err != nil && !e.IsRecordNotFound(err) {
			return nil, e.New(e.DBError, err)
		}
		if user != nil {
			resp.CreatorName = user.Name
		}
		resp.CreatedTime = time.Now().Unix() - resp.CreatedAt.Unix()
		if resp.EndAt != nil {
			resp.EndTime = time.Now().Unix() - resp.EndAt.Unix()
		}

	}

	return page.PageResp{
		Total:    p.MustTotal(),
		PageSize: p.Size,
		List:     taskResp,
	}, nil
}

type DetailTaskResp struct {
	models.Task

	OrgId       uint   `json:"orgId"`
	Description string `json:"description"`
	RepoId      string `json:"repoId"`
	RepoAddr    string `json:"repoAddr"`
	RepoBranch  string `json:"repoBranch"`
	SaveState   bool   `json:"saveState"`
	Varfile     string `json:"varfile"`
	Extra       string `json:"extra"`
	CreatorName string `json:"creatorName"`
}

func DetailTask(c *ctx.ServiceCtx, form *forms.DetailTaskForm) (interface{}, e.Error) {
	resp := DetailTaskResp{}
	tx := c.DB().Debug()
	if err := services.TaskDetail(tx, form.TaskId).
		First(&resp); err != nil {
		return nil, e.New(e.DBError, err)
	}
	user, err := services.GetUserById(tx, resp.Creator)
	if err != nil && !e.IsRecordNotFound(err) {
		return nil, e.New(e.DBError, err)
	}
	if user != nil {
		resp.CreatorName = user.Name
	}
	return resp, nil
}

func CreateTask(c *ctx.ServiceCtx, form *forms.CreateTaskForm) (interface{}, e.Error) {
	guid := utils.GenGuid("run")

	logPath := filepath.Join(form.TemplateGuid, guid, consts.TaskLogName)
	backend := models.TaskBackendInfo{
		BackendUrl:  fmt.Sprintf("http://%s:%d/api/v1", form.CtServiceIp, form.CtServicePort),
		CtServiceId: form.CtServiceId,
		LogFile:     logPath,
	}

	tpl, err := services.GetTemplateByGuid(c.DB(), form.TemplateGuid)
	if err != nil {
		return nil, err
	}
	vcs, er := services.QueryVcsByVcsId(tpl.VcsId, c.DB())
	if er != nil {
		return nil, er
	}
	var commitId string
	if vcs.VcsType == consts.GitTypeGitLab {
		git, err := vcs2.GetGitConn(vcs.VcsToken, vcs.Address)
		if err != nil {
			return nil, err
		}
		commits, _, commitErr := git.Commits.ListCommits(tpl.RepoId, &gitlab.ListCommitsOptions{})
		if commitErr != nil {
			return nil, e.New(e.GitLabError, commitErr)
		}

		if commits != nil {
			commitId = commits[0].ID
		}
	}

	if vcs.VcsType == consts.GitTypeGitEA {
		commit, err := vcs2.GetGiteaBranchCommitId(vcs, tpl.RepoId, tpl.RepoBranch)
		if err != nil {
			return nil, e.New(e.GitLabError, fmt.Errorf("query commit id error: %v", er))
		}
		commitId = commit
	}

	task, err := services.CreateTask(c.DB(), models.Task{
		TemplateId:   form.TemplateId,
		TemplateGuid: form.TemplateGuid,
		Guid:         guid,
		TaskType:     form.TaskType,
		Status:       consts.TaskPending,
		Creator:      c.UserId,
		Name:         form.Name,
		BackendInfo:  &backend,
		CtServiceId:  form.CtServiceId,
		CommitId:     commitId,
	})
	if err != nil {
		return nil, err
	}

	//发送通知
	go services.SendMail(c.DB(), c.OrgId, task)

	return task, nil
}

type LastTaskResp struct {
	models.Task
	CreatorName string `json:"creatorName" form:"creatorName" `
	RepoBranch  string `json:"repoBranch" form:"repoBranch" `
}

func LastTask(c *ctx.ServiceCtx, form *forms.LastTaskForm) (interface{}, e.Error) {
	tx := c.DB().Debug()
	taskResp := LastTaskResp{}
	tpl, err := services.GetTemplateById(tx, form.TemplateId)
	if err != nil {
		return nil, err
	}
	if err := services.LastTask(tx, form.TemplateId).Scan(&taskResp); err != nil && err != gorm.ErrRecordNotFound {
		return nil, e.New(e.DBError, err)
	}
	if taskResp.Creator != 0 {
		user, err := services.GetUserById(tx, taskResp.Creator)
		if err != nil && !e.IsRecordNotFound(err) {
			return nil, err
		}
		if user != nil {
			taskResp.CreatorName = user.Name
		}
	}
	taskResp.RepoBranch = tpl.RepoBranch
	return taskResp, nil
}

func TaskStateList(c *ctx.ServiceCtx, form *forms.TaskStateListForm) (interface{}, e.Error) {
	query := c.DB()
	task, err := services.GetTaskByGuid(query.Debug(), form.TaskGuid)
	if err != nil && !e.IsRecordNotFound(err){
		return nil, err
	}

	return services.TaskStateList(task)
}
