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

	query, _ := services.QueryTemplate(c.DB().Debug(), form.Status, form.Q, form.TaskStatus, statusList, c.OrgId)

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
			OrgId:                  c.OrgId,
			Name:                   form.Name,
			Guid:                   guid,
			Description:            form.Description,
			RepoId:                 form.RepoId,
			RepoBranch:             form.RepoBranch,
			RepoAddr:               form.RepoAddr,
			SaveState:              *form.SaveState,
			Vars:                   models.JSON(string(jsons)),
			Varfile:                form.Varfile,
			Extra:                  form.Extra,
			Timeout:                form.Timeout,
			Creator:                c.UserId,
			VcsId:                  form.VcsId,
			DefaultRunnerAddr:      form.DefaultRunnerAddr,
			DefaultRunnerPort:      form.DefaultRunnerPort,
			DefaultRunnerServiceId: form.DefaultRunnerServiceId,
			Playbook:               form.Playbook,
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

func UpdateTemplate(c *ctx.ServiceCtx, form *forms.UpdateTemplateForm) (*models.Template, e.Error) {
	c.AddLogField("action", fmt.Sprintf("update template %d", form.Id))
	vars := make([]forms.Var, 0)
	newVars := make(map[string]string, 0)
	tpl, err := services.GetTemplateById(c.DB(), form.Id)
	if err != nil {
		return nil, err
	}
	if !tpl.Vars.IsNull() {
		_ = json.Unmarshal(tpl.Vars, &vars)
	}

	for _, v := range vars {
		newVars[v.Id] = v.Value
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

	if form.HasKey("defaultRunnerServiceId") {
		attrs["defaultRunnerServiceId"] = form.DefaultRunnerServiceId
	}

	if form.HasKey("defaultRunnerPort") {
		attrs["defaultRunnerPort"] = form.DefaultRunnerPort
	}

	if form.HasKey("defaultRunnerAddr") {
		attrs["defaultRunnerAddr"] = form.DefaultRunnerAddr
	}

	if form.HasKey("playbook") {
		attrs["playbook"] = form.Playbook
	}

	if form.HasKey("vars") {
		vars := form.Vars
		for index, v := range vars {
			if *v.IsSecret && v.Value != "" {
				encryptedValue, err := utils.AesEncrypt(v.Value)
				vars[index].Value = encryptedValue
				if err != nil {
					return nil, nil
				}
			}
			if v.Value == "" && *v.IsSecret {
				vars[index].Value = newVars[v.Id]
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

	return services.UpdateTemplate(c.DB().Debug(), form.Id, attrs)
}

func DetailTemplate(c *ctx.ServiceCtx, form *forms.DetailTemplateForm) (interface{}, e.Error) {
	vars := make([]forms.Var, 0)
	newVars := make([]forms.Var, 0)
	tpl, err := services.DetailTemplate(c.DB(), form.Id)
	if err != nil {
		return nil, err
	}
	if !tpl.Vars.IsNull() {
		_ = json.Unmarshal(tpl.Vars, &vars)
	}
	for _, v := range vars {
		if *v.IsSecret {
			newVars = append(newVars, forms.Var{
				Key:         v.Key,
				Value:       "",
				IsSecret:    v.IsSecret,
				Type:        v.Type,
				Description: v.Description,
				Id:          v.Id,
			})
		} else {
			newVars = append(newVars, v)
		}
	}
	b, _ := json.Marshal(newVars)

	tpl.Vars = models.JSON(b)

	return tpl, nil
}

type OverviewTemplateResp struct {
	RepoAddr               string    `json:"repoAddr" form:"repoAddr" `
	RepoBranch             string    `json:"repoBranch" form:"repoBranch" `
	Name                   string    `json:"name" form:"name" `
	Guid                   string    `json:"guid" form:"guid" `
	TaskPlanCount          int64     `json:"taskPlanCount" form:"taskPlanCount" `
	TaskApplyCount         int64     `json:"taskApplyCount" form:"taskApplyCount" `
	TaskAvgPlanTime        int64     `json:"taskAvgPlanTime" form:"taskAvgPlanTime" `
	TaskAvgApplyTime       int64     `json:"taskAvgApplyTime" form:"taskAvgApplyTime" `
	TaskPlanFailedPercent  float64   `json:"taskPlanFailedPercent" form:"taskPlanFailedPercent" `
	TaskApplyFailedPercent float64   `json:"taskApplyFailedPercent" form:"taskApplyFailedPercent" `
	TaskPlanFailedCount    float64   `json:"taskPlanFailedCount" form:"taskPlanFailedCount" `
	TaskApplyFailedCount   float64   `json:"taskApplyFailedCount" form:"taskApplyFailedCount" `
	ActiveCreatorName      []string  `json:"activeCreatorName" form:"activeCreatorName" `
	Task                   []Task    `json:"task" form:"task" `
	TaskLastUpdatedAt      time.Time `json:"taskLastUpdatedAt" form:"taskLastUpdatedAt" `
	CreatorName            string    `json:"creatorName" form:"creatorName" `
}
type Task struct {
	Name        string    `json:"name" form:"name" `
	Status      string    `json:"status" form:"status" `
	Guid        string    `json:"guid" form:"guid" `
	TaskType    string    `json:"taskType" form:"taskType" `
	CreatedAt   time.Time `json:"createdAt" form:"createdAt" `
	EndAt       time.Time `json:"endAt" form:"endAt" `
	CreatorName string    `json:"creatorName" form:"creatorName" `
	CreatedTime int64     `json:"createdTime" form:"createdTime" `
	EndTime     int64     `json:"endTime" form:"endTime" `
	CommitId    string    `json:"commitId" gorm:"null;comment:'COMMIT ID'"`
	Id          uint      `json:"id" form:"id" `
	Add         string    `json:"add" gorm:"default:0"`
	Change      string    `json:"change" gorm:"default:0"`
	Destroy     string    `json:"destroy" gorm:"default:0"`
	AllowApply  bool      `json:"allowApply" gorm:"default:false"`
	RepoBranch  string    `json:"repoBranch" form:"repoBranch" `
	CtServiceId string    `json:"ctServiceId" form:"ctServiceId" `
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
		taskLastUpdatedAt      time.Time
	)

	if err := services.OverviewTemplate(tx, form.Id).First(&tpl); err != nil {
		return nil, e.New(e.DBError, err)
	}

	tasks, err := services.OverviewTemplateTask(tx, form.Id)
	if err != nil {
		return nil, e.New(e.DBError, err)
	}

	for index, task := range tasks {
		if index == 0 {
			taskLastUpdatedAt = task.CreatedAt
		}
		user, _ := services.GetUserById(tx, task.Creator)

		if task.TaskType == consts.TaskPlan {
			if task.Status == consts.TaskFailed || task.Status == consts.TaskTimeout {
				taskPlanFailedCount++
			}
			taskPlanCount++
			taskAvgPlanTimeCount += task.EndAt.Unix() - task.StartAt.Unix()
		}

		if task.TaskType == consts.TaskApply {
			if task.Status == consts.TaskFailed || task.Status == consts.TaskTimeout {
				taskApplyFailedCount++
			}
			taskApplyCount++
			taskAvgApplyTimeCount += task.EndAt.Unix() - task.StartAt.Unix()
		}

		// timeout也算失败
		if user != nil {
			activeCreatorName = append(activeCreatorName, user.Name)
		}
		//取最新的3个任务
		if index < 3 {
			var username string
			if user != nil {
				username = user.Name
			}
			taskList = append(taskList, Task{
				Name:        task.Name,
				Status:      task.Status,
				Guid:        task.Guid,
				TaskType:    task.TaskType,
				CreatedAt:   task.CreatedAt,
				CreatorName: username,
				CommitId:    task.CommitId,
				CreatedTime: time.Now().Unix() - task.CreatedAt.Unix(),
				EndTime:     time.Now().Unix() - task.EndAt.Unix(),
				Id:          task.Id,
				Add:         task.Add,
				Destroy:     task.Destroy,
				Change:      task.Change,
				AllowApply:  task.AllowApply,
				RepoBranch:  tpl.RepoBranch,
				CtServiceId: task.CtServiceId,
				EndAt:       *task.EndAt,
			})
		}
	}
	if taskPlanCount > 0 {
		taskPlanFailedPercent = taskPlanFailedCount / float64(taskPlanCount) * 100
		if taskAvgPlanTimeCount > 0 {
			taskAvgApplyTime = taskApplyCount / taskAvgPlanTimeCount
		}
	}

	if taskApplyCount > 0 {
		taskApplyFailedPercent = taskApplyFailedCount / float64(taskApplyCount) * 100
		if taskAvgApplyTimeCount > 0 {
			taskAvgApplyTime = taskApplyCount / taskAvgApplyTimeCount
		}
	}
	user, _ := services.GetUserById(tx, tpl.Creator)
	var name string
	if user != nil {
		name = user.Name
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
		TaskApplyFailedCount:   taskApplyFailedCount,
		TaskPlanFailedCount:    taskPlanFailedCount,
		ActiveCreatorName:      utils.RemoveDuplicateElement(activeCreatorName),
		Task:                   taskList,
		TaskLastUpdatedAt:      taskLastUpdatedAt,
		CreatorName:            name,
	}, nil
}
