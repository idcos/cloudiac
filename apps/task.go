package apps

import (
	"cloudiac/configs"
	"cloudiac/consts"
	"cloudiac/consts/e"
	"cloudiac/libs/ctx"
	"cloudiac/libs/page"
	"cloudiac/models"
	"cloudiac/models/forms"
	"cloudiac/services"
	"cloudiac/utils"
	"encoding/json"
	"fmt"
	"os"
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
		if err != nil {
			return nil, e.New(e.DBError, err)
		}
		resp.CreatorName = user.Name
		resp.CreatedTime = time.Now().Unix() - resp.CreatedAt.Unix()
		resp.EndTime = time.Now().Unix() - resp.EndAt.Unix()
	}

	return page.PageResp{
		Total:    p.MustTotal(),
		PageSize: p.Size,
		List:     taskResp,
	}, nil
}

type DetailTaskResp struct {
	models.Task

	OrgId       uint   `json:"orgId" gorm:"size:32;not null;comment:'组织ID'"`
	Description string `json:"description" gorm:"size:255;comment:'描述'"`
	RepoId      int    `json:"repoId" gorm:"size:32;comment:'仓库ID'"`
	RepoAddr    string `json:"repoAddr" gorm:"size:128;default:'';comment:'仓库地址'"`
	RepoBranch  string `json:"repoBranch" gorm:"size:64;default:'master';comment:'仓库分支'"`
	SaveState   *bool  `json:"saveState" gorm:"defalut:false;comment:'是否保存状态'"`
	Varfile     string `json:"varfile" gorm:"size:128;default:'';comment:'变量文件'"`
	Extra       string `json:"extra" gorm:"size:128;default:'';comment:'附加信息'"`
	CreatorName string `json:"creatorName" form:"creatorName" `
}

func DetailTask(c *ctx.ServiceCtx, form *forms.DetailTaskForm) (interface{}, e.Error) {
	resp := DetailTaskResp{}
	tx := c.DB().Debug()
	if err := services.TaskDetail(tx, form.TaskId).
		First(&resp); err != nil {
		return nil, e.New(e.DBError, err)
	}
	user, err := services.GetUserById(tx, resp.Creator)
	if err != nil {
		return nil, e.New(e.DBError, err)
	}
	resp.CreatorName = user.Name
	return resp, nil
}

func CreateTask(c *ctx.ServiceCtx, form *forms.CreateTaskForm) (interface{}, e.Error) {
	guid := utils.GenGuid("run")
	conf := configs.Get()
	logPath := fmt.Sprintf("%s/%s/%s", conf.Task.LogPath, form.TemplateGuid, guid)
	b, _ := json.Marshal(map[string]interface{}{
		"backend_url": fmt.Sprintf("http://%s:%d/api/v1", form.CtServiceIp, form.CtServicePort),
		"ctServiceId": form.CtServiceId,
		"log_file":    logPath,
		"log_offset":  0,
	})

	if err := os.MkdirAll(logPath, os.ModePerm); err != nil {
		return nil, e.New(e.IOError, err)
	}
	task, err := services.CreateTask(c.DB().Debug(), models.Task{
		TemplateId:   form.TemplateId,
		TemplateGuid: form.TemplateGuid,
		Guid:         guid,
		TaskType:     form.TaskType,
		Status:       consts.TaskPending,
		Creator:      c.UserId,
		Name:         form.Name,
		BackendInfo:  models.JSON(b),
		CtServiceId:  form.CtServiceId,
	})
	if err != nil {
		return nil, err
	}
	//todo Task数量够多的情况下需要引入第三方组件
	go services.RunTaskToRunning(task, c.DB(), c.MustOrg().Guid)
	return task, nil
}

type LastTaskResp struct {
	models.Task
	CreatorName string `json:"creatorName" form:"creatorName" `
}

func LastTask(c *ctx.ServiceCtx, form *forms.LastTaskForm) (interface{}, e.Error) {
	tx := c.DB().Debug()
	taskResp := LastTaskResp{}

	if err := services.LastTask(tx, form.TemplateId).Scan(&taskResp); err != nil {
		return nil, e.New(e.DBError, err)
	}

	user, err := services.GetUserById(tx, taskResp.Creator)
	if err != nil {
		return nil, err
	}
	taskResp.CreatorName = user.Name
	return taskResp, nil
}
