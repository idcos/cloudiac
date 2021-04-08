package apps

import (
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
	"strings"
	"time"
)

type SearchTemplateResp struct {
	Id            uint      `json:"id"`
	Name          string    `json:"name"`
	Guid          string    `json:"guid"`
	RepoAddr      string    `json:"repoAddr"`
	TaskStatus    string    `json:"taskStatus"`
	TaskGuid      string    `json:"taskGuid"`
	TaskUpdatedAt time.Time `json:"taskUpdatedAt"`
}

func SearchTemplate(c *ctx.ServiceCtx, form *forms.SearchTemplateForm) (interface{}, e.Error) {
	statusList := make([]string, 0)
	if form.TaskStatus == "all" || form.TaskStatus == "" {
		statusList = append(statusList, []string{
			"pending",
			"running",
			"failed",
			"complete",
			"timeout",
		}...)
	} else {
		statusList = strings.Split(form.TaskStatus, ",")
	}

	query, _ := services.QueryTemplate(c.DB().Debug(), form.Status, form.Q,form.TaskStatus ,statusList)

	p := page.New(form.CurrentPage(), form.PageSize(), query)
	templates := make([]*SearchTemplateResp, 0)
	if err := p.Scan(&templates); err != nil {
		return nil, e.New(e.DBError, err)
	}

	return page.PageResp{
		Total:    p.MustTotal(),
		PageSize: p.Size,
		List:     templates,
	}, nil
}

func CreateTemplate(c *ctx.ServiceCtx, form *forms.CreateTemplateForm) (*models.Template, e.Error) {
	c.AddLogField("action", fmt.Sprintf("create template %s", form.Name))

	tx := c.Tx().Debug()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	guid := utils.GenGuid("ct")
	template, err := func() (*models.Template, e.Error) {
		var (
			template *models.Template
			err      e.Error
		)

		vars := form.Vars
		for index, v := range vars {
			if *v.IsSecret {
				encryptedValue, err := utils.AesEncrypt(v.Value)
				vars[index].Value = encryptedValue
				if err != nil {
					return nil, nil
				}
			}
		}
		jsons, _ := json.Marshal(vars)

		template, err = services.CreateTemplate(tx, models.Template{
			OrgId:       c.OrgId,
			Name:        form.Name,
			Guid:        guid,
			Description: form.Description,
			RepoId:      form.RepoId,
			RepoBranch:  form.RepoBranch,
			RepoAddr:    form.RepoAddr,
			SaveState:   *form.SaveState,
			Vars:        models.JSON(string(jsons)),
			Varfile:     form.Varfile,
			Extra:       form.Extra,
			Timeout:     form.Timeout,
			Creator:     c.UserId,
		})
		if err != nil {
			return nil, err
		}

		return template, nil
	}()
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, e.New(e.DBError, err)
	}

	return template, nil
}

func UpdateTemplate(c *ctx.ServiceCtx, form *forms.UpdateTemplateForm) (user *models.Template, err e.Error) {
	c.AddLogField("action", fmt.Sprintf("update template %d", form.Id))
	if form.Id == 0 {
		return nil, e.New(e.BadRequest, fmt.Errorf("missing 'id'"))
	}

	attrs := models.Attrs{}
	if form.HasKey("name") {
		attrs["name"] = form.Name
	}

	if form.HasKey("description") {
		attrs["description"] = form.Description
	}

	if form.HasKey("saveState") {
		attrs["saveState"] = form.SaveState
	}

	if form.HasKey("vars") {
		vars := form.Vars
		for index, v := range vars {
			if *v.IsSecret {
				encryptedValue, err := utils.AesEncrypt(v.Value)
				vars[index].Value = encryptedValue
				if err != nil {
					return nil, nil
				}
			}
		}
		jsons, _ := json.Marshal(vars)
		attrs["vars"] = jsons
	}

	if form.HasKey("varfile") {
		attrs["varfile"] = form.Varfile
	}

	if form.HasKey("extra") {
		attrs["extra"] = form.Extra
	}

	if form.HasKey("timeout") {
		attrs["timeout"] = form.Timeout
	}

	if form.HasKey("status") {
		attrs["status"] = form.Status
	}

	user, err = services.UpdateTemplate(c.DB(), form.Id, attrs)
	return
}

func DetailTemplate(c *ctx.ServiceCtx, form *forms.DetailTemplateForm) (interface{}, e.Error) {
	return services.DetailTemplate(c.DB(), form.Id)
}

type OverviewTemplateResp struct {
	RepoAddr               string   `json:"repoAddr" form:"repoAddr" `
	RepoBranch             string   `json:"repoBranch" form:"repoBranch" `
	Name                   string   `json:"name" form:"name" `
	Guid                   string   `json:"guid" form:"guid" `
	TaskPlanCount          int64    `json:"taskPlanCount" form:"taskPlanCount" `
	TaskApplyCount         int64    `json:"taskApplyCount" form:"taskApplyCount" `
	TaskAvgPlanTime        int64    `json:"taskAvgPlanTime" form:"taskAvgPlanTime" `
	TaskAvgApplyTime       int64    `json:"taskAvgApplyTime" form:"taskAvgApplyTime" `
	TaskPlanFailedPercent  float64  `json:"taskPlanFailedPercent" form:"taskPlanFailedPercent" `
	TaskApplyFailedPercent float64  `json:"taskApplyFailedPercent" form:"taskApplyFailedPercent" `
	ActiveCreatorName      []string `json:"activeCreatorName" form:"activeCreatorName" `
	Task                   []Task   `json:"task" form:"task" `
}
type Task struct {
	Name        string    `json:"name" form:"name" `
	Status      string    `json:"status" form:"status" `
	Guid        string    `json:"guid" form:"guid" `
	TaskType    string    `json:"taskType" form:"taskType" `
	CreateAt    time.Time `json:"createAt" form:"createAt" `
	CreatorName string    `json:"creatorName" form:"creatorName" `
}

func OverviewTemplate(c *ctx.ServiceCtx, form *forms.OverviewTemplateForm) (interface{}, e.Error) {
	tx := c.DB().Debug()
	tpl := models.Template{}
	taskList := make([]Task, 0)
	activeCreatorName := make([]string, 0)
	var (
		taskPlanCount          int64
		taskApplyCount         int64
		taskAvgPlanTimeCount   int64
		taskAvgApplyTimeCount  int64
		taskPlanFailedCount    float64
		taskApplyFailedCount   float64
		taskAvgPlanTime        int64
		taskAvgApplyTime       int64
		taskPlanFailedPercent  float64
		taskApplyFailedPercent float64
	)

	if err := services.OverviewTemplate(tx, form.Id).First(&tpl); err != nil {
		return nil, e.New(e.DBError, err)
	}

	tasks, err := services.OverviewTemplateTask(tx, form.Id)
	if err != nil {
		return nil, e.New(e.DBError, err)
	}

	for _, task := range tasks {
		user, err := services.GetUserById(tx, task.Creator)
		if err != nil {
			return nil, e.New(e.DBError, err)
		}
		if task.TaskType == consts.TaskPlan {
			if task.Status == consts.TaskFailed || task.Status == consts.TaskTimeoout {
				taskPlanFailedCount++
			}
			taskPlanCount++
			taskAvgPlanTimeCount += task.EndAt.Unix() - task.StartAt.Unix()
		}

		if task.TaskType == consts.TaskApply {
			if task.Status == consts.TaskFailed || task.Status == consts.TaskTimeoout {
				taskApplyFailedCount++
			}
			taskApplyCount++
			taskAvgApplyTimeCount += task.EndAt.Unix() - task.StartAt.Unix()
		}

		// timeout也算失败

		activeCreatorName = append(activeCreatorName, user.Name)

		taskList = append(taskList, Task{
			Name:        task.Name,
			Status:      task.Status,
			Guid:        task.Guid,
			TaskType:    task.TaskType,
			CreateAt:    task.CreatedAt,
			CreatorName: user.Name,
		})

	}
	if taskPlanCount > 0 {
		taskPlanFailedPercent = taskPlanFailedCount / float64(taskPlanCount)
		if taskAvgPlanTimeCount > 0 {
			taskAvgApplyTime = taskApplyCount / taskAvgPlanTimeCount
		}
	}

	if taskApplyCount > 0 {
		taskApplyFailedPercent = taskApplyFailedCount / float64(taskApplyCount)
		if taskAvgApplyTimeCount > 0 {
			taskAvgApplyTime = taskApplyCount / taskAvgApplyTimeCount
		}
	}

	return OverviewTemplateResp{
		RepoAddr:               tpl.RepoAddr,
		RepoBranch:             tpl.RepoBranch,
		Name:                   tpl.Name,
		Guid:                   tpl.Guid,
		TaskPlanCount:          taskPlanCount,
		TaskApplyCount:         taskApplyCount,
		TaskAvgPlanTime:        taskAvgPlanTime,
		TaskAvgApplyTime:       taskAvgApplyTime,
		TaskPlanFailedPercent:  taskPlanFailedPercent,
		TaskApplyFailedPercent: taskApplyFailedPercent,
		ActiveCreatorName:      activeCreatorName,
		Task:                   taskList,
	}, nil
}
