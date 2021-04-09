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
)

type SearchTaskResp struct {
	models.Task
}

func SearchTask(c *ctx.ServiceCtx, form *forms.SearchTaskForm) (interface{}, e.Error) {
	query := services.QueryTask(c.DB())
	query = query.Where("template_id = ?", form.TemplateId)
	if form.Status != "" {
		query = query.Where("status = ?", form.Status)
	}
	if form.Q != "" {
		qs := "%" + form.Q + "%"
		query = query.Where("name LIKE ? OR description LIKE ?", qs, qs)
	}

	query = query.Order("created_at DESC")
	p := page.New(form.CurrentPage(), form.PageSize(), query)
	tasks := make([]*models.Task, 0)
	if err := p.Scan(&tasks); err != nil {
		return nil, e.New(e.DBError, err)
	}

	return page.PageResp{
		Total:    p.MustTotal(),
		PageSize: p.Size,
		List:     tasks,
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
}

func DetailTask(c *ctx.ServiceCtx, form *forms.DetailTaskForm) (interface{}, e.Error) {
	resp := DetailTaskResp{}
	if err := services.TaskDetail(c.DB().Debug(), form.TaskId).
		First(&resp); err != nil {
		return nil, e.New(e.DBError, err)
	}
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

	return services.CreateTask(c.DB().Debug(), models.Task{
		TemplateId:   form.TemplateId,
		TemplateGuid: form.TemplateGuid,
		Guid:         guid,
		TaskType:     form.TaskType,
		Status:       consts.TaskPending,
		Creator:      c.UserId,
		Name:         form.Name,
		BackendInfo:  models.JSON(b),
		CtServiceId:  form.CtServiceId,
		Timeout:      form.Timeout,
	})
}

func LastTask(c *ctx.ServiceCtx, form *forms.LastTaskForm) (interface{}, e.Error) {
	return services.LastTask(c.DB().Debug(), form.TemplateId)
}
