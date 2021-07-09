package apps

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
	"fmt"
	"net/http"
)

// CreateTask 创建任务
func CreateTask(c *ctx.ServiceCtx, form *forms.CreateTaskForm) (*models.Task, e.Error) {
	c.AddLogField("action", fmt.Sprintf("create task %s", form.Name))

	// TODO
	task, err := services.CreateTask(c.DB(), &models.Env{Name: "env"}, models.Task{
		Name:      form.Name,
		CreatorId: c.UserId,
	})
	if err != nil && err.Code() == e.TaskAlreadyExists {
		return nil, e.New(err.Code(), err, http.StatusBadRequest)
	} else if err != nil {
		c.Logger().Errorf("error creating task, err %s", err)
		return nil, e.AutoNew(err, e.DBError)
	}
	return task, nil
}

// SearchTask 任务查询
func SearchTask(c *ctx.ServiceCtx, form *forms.SearchTaskForm) (interface{}, e.Error) {
	query := services.QueryTask(c.DB())
	// 默认按创建时间逆序排序
	if form.SortField() == "" {
		query = query.Order("created_at DESC")
	}
	rs, err := getPage(query, form, &models.Task{})
	if err != nil {
		c.Logger().Errorf("error get page, err %s", err)
	}
	return rs, err
}

type taskDetailResp struct {
	models.Task
	Creator string `json:"creator" example:"超级管理员"`
}

// TaskDetail 任务信息详情
func TaskDetail(c *ctx.ServiceCtx, form forms.DetailTaskForm) (*taskDetailResp, e.Error) {
	taskIds, er := services.GetOrgIdsByUser(c.DB(), c.UserId)
	if er != nil {
		c.Logger().Errorf("error get task id by user, err %s", er)
		return nil, e.New(e.DBError, er)
	}
	if form.Id.InArray(taskIds...) == false && c.IsSuperAdmin == false {
		// 请求了一个不存在的 task，因为 task id 是在 path 传入，这里我们返回 404
		return nil, e.New(e.TaskNotExists, http.StatusNotFound)
	}

	var (
		task *models.Task
		user *models.User
		err  e.Error
	)
	task, err = services.GetTaskById(c.DB(), form.Id)
	if err != nil && err.Code() == e.TaskNotExists {
		return nil, e.New(e.TaskNotExists, err, http.StatusNotFound)
	} else if err != nil {
		c.Logger().Errorf("error get task by id, err %s", err)
		return nil, e.New(e.DBError, err)
	}
	user, err = services.GetUserById(c.DB(), task.CreatorId)
	if err != nil && err.Code() == e.UserNotExists {
		// 报 500 错误，正常情况用户不应该找不到，除非被意外删除
		return nil, e.New(e.UserNotExists, err, http.StatusInternalServerError)
	} else if err != nil {
		c.Logger().Errorf("error get user by id, err %s", err)
		return nil, e.New(e.DBError, err)
	}

	var o = taskDetailResp{
		Task:    *task,
		Creator: user.Name,
	}

	return &o, nil
}

// CurrentTask 当前任务信息
func CurrentTask(c *ctx.ServiceCtx, form *forms.CurrentTaskForm) (*taskDetailResp, e.Error) {
	if c.OrgId == "" || c.ProjectId == "" {
		return nil, e.New(e.BadRequest, http.StatusBadRequest)
	}
	query := c.DB().Where("iac_env.org_id = ? AND iac_env.project_id = ?", c.OrgId, c.ProjectId)
	env, err := services.GetEnvById(query, form.Id)
	if err != nil && err.Code() == e.EnvNotExists {
		return nil, e.New(err.Code(), err, http.StatusNotFound)
	} else if err != nil {
		c.Logger().Errorf("error get task by id, err %s", err)
		return nil, e.New(e.DBError, err)
	}

	// FIXME: 环境处于非活跃状态，没有任何在执行的任务？
	if env.CurrentTaskId == "" {
		return nil, nil
	}

	task, err := services.GetTaskById(c.DB(), form.Id)
	if err != nil && err.Code() == e.TaskNotExists {
		return nil, e.New(e.TaskNotExists, err, http.StatusNotFound)
	} else if err != nil {
		c.Logger().Errorf("error get task by id, err %s", err)
		return nil, e.New(e.DBError, err)
	}
	user, err := services.GetUserById(c.DB(), task.CreatorId)
	if err != nil && err.Code() == e.UserNotExists {
		// 报 500 错误，正常情况用户不应该找不到，除非被意外删除
		return nil, e.New(e.UserNotExists, err, http.StatusInternalServerError)
	} else if err != nil {
		c.Logger().Errorf("error get user by id, err %s", err)
		return nil, e.New(e.DBError, err)
	}

	var t = taskDetailResp{
		Task:    *task,
		Creator: user.Name,
	}

	return &t, nil
}

// ApproveTask 审批执行计划
func ApproveTask(c *ctx.ServiceCtx, form *forms.ApproveTaskForm) (interface{}, e.Error) {
	c.AddLogField("action", fmt.Sprintf("approve task %s", form.Id))

	if c.OrgId == "" || c.ProjectId == "" {
		return nil, e.New(e.BadRequest, http.StatusBadRequest)
	}
	query := c.DB().Where("iac_env.org_id = ? AND iac_env.project_id = ?", c.OrgId, c.ProjectId)
	task, err := services.GetTask(query, form.Id)
	if err != nil && err.Code() != e.TaskNotExists {
		return nil, e.New(err.Code(), err, http.StatusNotFound)
	} else if err != nil {
		c.Logger().Errorf("error get env, err %s", err)
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	}

	// FIXME
	if task.Status != models.TaskPending {
		return nil, e.New(e.TaskApproveNotPending, http.StatusBadRequest)
	}

	// TODO 发出审批通过/驳回信号

	return nil, nil
}
