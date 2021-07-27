package apps

import (
	"bufio"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/libs/page"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
	"fmt"
	"github.com/gin-contrib/sse"
	"io"
	"net/http"
	"path"
	"strconv"
)

// SearchTask 任务查询
func SearchTask(c *ctx.ServiceCtx, form *forms.SearchTaskForm) (interface{}, e.Error) {
	query := services.QueryTask(c.DB())
	if form.EnvId != "" {
		query = query.Where("env_id = ?", form.EnvId)
	}
	// 默认按创建时间逆序排序
	if form.SortField() == "" {
		query = query.Order("created_at DESC")
	}

	p := page.New(form.CurrentPage(), form.PageSize(), query)
	details := make([]*taskDetailResp, 0)
	if err := p.Scan(&details); err != nil {
		return nil, e.New(e.DBError, err)
	}

	if details != nil {
		for _, env := range details {
			// 隐藏敏感字段
			env.HideSensitiveVariable()
		}
	}

	return page.PageResp{
		Total:    p.MustTotal(),
		PageSize: p.Size,
		List:     details,
	}, nil
}

type taskDetailResp struct {
	models.Task
	Creator string `json:"creator" example:"超级管理员"`
}

// TaskDetail 任务信息详情
func TaskDetail(c *ctx.ServiceCtx, form forms.DetailTaskForm) (*taskDetailResp, e.Error) {
	orgIds, er := services.GetOrgIdsByUser(c.DB(), c.UserId)
	if er != nil {
		c.Logger().Errorf("error get task id by user, err %s", er)
		return nil, e.New(e.DBError, er)
	}
	if c.OrgId.InArray(orgIds...) == false && c.IsSuperAdmin == false {
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
		user = &models.User{}
		c.Logger().Errorf("task creator '%s' not exists", task.CreatorId)
	} else if err != nil {
		c.Logger().Errorf("error get user by id, err %s", err)
		return nil, e.New(e.DBError, err)
	}

	// 隐藏敏感字段
	task.HideSensitiveVariable()
	var o = taskDetailResp{
		Task:    *task,
		Creator: user.Name,
	}

	return &o, nil
}

// LastTask 最新任务信息
func LastTask(c *ctx.ServiceCtx, form *forms.LastTaskForm) (*taskDetailResp, e.Error) {
	if c.OrgId == "" || c.ProjectId == "" {
		return nil, e.New(e.BadRequest, http.StatusBadRequest)
	}
	query := c.DB().Where("org_id = ? AND project_id = ?", c.OrgId, c.ProjectId)
	env, err := services.GetEnvById(query, form.Id)
	if err != nil && err.Code() == e.EnvNotExists {
		return nil, e.New(err.Code(), err, http.StatusNotFound)
	} else if err != nil {
		c.Logger().Errorf("error get task by id, err %s", err)
		return nil, e.New(e.DBError, err)
	}

	// 环境处于非活跃状态，没有任何在执行的任务
	if env.LastTaskId == "" {
		return nil, nil
	}

	task, err := services.GetTaskById(query, env.LastTaskId)
	if err != nil && err.Code() == e.TaskNotExists {
		return nil, e.New(e.TaskNotExists, err, http.StatusNotFound)
	} else if err != nil {
		c.Logger().Errorf("error get task by id, err %s", err)
		return nil, e.New(e.DBError, err)
	}
	user, err := services.GetUserById(c.DB(), task.CreatorId)
	if err != nil && err.Code() == e.UserNotExists {
		user = &models.User{}
		c.Logger().Errorf("task creator '%s' not exists", task.CreatorId)
	} else if err != nil {
		c.Logger().Errorf("error get user by id, err %s", err)
		return nil, e.New(e.DBError, err)
	}

	// 隐藏敏感字段
	task.HideSensitiveVariable()
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

	taskQuery := services.QueryWithProjectId(services.QueryWithOrgId(c.DB(), c.OrgId), c.ProjectId)
	task, err := services.GetTask(taskQuery, form.Id)
	if err != nil && err.Code() != e.TaskNotExists {
		return nil, e.New(err.Code(), err, http.StatusNotFound)
	} else if err != nil {
		c.Logger().Errorf("error get task, err %s", err)
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	}

	if task.Status != models.TaskApproving {
		return nil, e.New(e.TaskApproveNotPending, http.StatusBadRequest)
	}

	step, err := services.GetTaskStep(c.DB(), task.Id, task.CurrStep)
	if err != nil && err.Code() == e.TaskStepNotExists {
		c.Logger().Errorf("task %s step %d not exist", task.Id, task.CurrStep, err)
		return nil, e.AutoNew(err, err.Code())
	} else if err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}

	// 己通过审批
	if step.IsApproved() || step.ApproverId != "" {
		return nil, e.New(e.TaskApproveNotPending, http.StatusBadRequest)
	}

	// 更新审批状态
	step.ApproverId = c.UserId
	switch form.Action {
	case forms.TaskActionApproved:
		err = services.ApproveTaskStep(c.DB(), task.Id, step.Index, c.UserId)
	case forms.TaskActionRejected:
		err = services.RejectTaskStep(c.DB(), task.Id, step.Index, c.UserId)
	}
	if err != nil {
		c.Logger().Errorf("error approve task, err %s", err)
		return nil, err
	}

	return nil, nil
}

func FollowTaskLog(c *ctx.GinRequestCtx, form forms.DetailTaskForm) e.Error {
	logger := c.Logger().WithField("func", "FollowTaskLog").WithField("taskId", form.Id)
	sc := c.ServiceCtx()
	rCtx := c.Context.Request.Context()

	// TODO 浏览器原生 SSE 实现不支持修改 header，所以这个接口暂时不作认证，待前端支持
	//task, er := services.GetTask(sc.ProjectDB(), form.Id)
	task, er := services.GetTask(sc.DB(), form.Id)
	if er != nil {
		logger.Errorf("get task: %v", er)
		if er.Code() == e.TaskNotExists {
			return e.New(er.Code(), http.StatusNotFound)
		}
		return er
	}

	pr, pw := io.Pipe()
	go func() {
		if err := services.FetchTaskLog(rCtx, task, pw); err != nil {
			logger.Errorf("fetch task log: %v", err)
		}
	}()

	scanner := bufio.NewScanner(pr)
	eventId := 0 // to indicate the message id
	for scanner.Scan() {
		c.Render(-1, sse.Event{
			Id:    strconv.Itoa(eventId),
			Event: "message",
			Data:  scanner.Text(),
		})
		c.Writer.Flush()
		eventId += 1
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		return e.New(e.InternalError, err)
	}
	return nil
}

// TaskOutput 任务Output信息详情
func TaskOutput(c *ctx.ServiceCtx, form forms.DetailTaskForm) (interface{}, e.Error) {
	orgIds, er := services.GetOrgIdsByUser(c.DB(), c.UserId)
	if er != nil {
		c.Logger().Errorf("error get task id by user, err %s", er)
		return nil, e.New(e.DBError, er)
	}
	if c.OrgId.InArray(orgIds...) == false && c.IsSuperAdmin == false {
		// 请求了一个不存在的 task，因为 task id 是在 path 传入，这里我们返回 404
		return nil, e.New(e.TaskNotExists, http.StatusNotFound)
	}

	var (
		task *models.Task
		err  e.Error
	)
	task, err = services.GetTaskById(c.DB(), form.Id)
	if err != nil && err.Code() == e.TaskNotExists {
		return nil, e.New(e.TaskNotExists, err, http.StatusNotFound)
	} else if err != nil {
		c.Logger().Errorf("error get task by id, err %s", err)
		return nil, e.New(e.DBError, err)
	}

	return task.Result.Outputs, nil
}

// SearchTaskResources 查询环境资源列表
func SearchTaskResources(c *ctx.ServiceCtx, form *forms.SearchTaskResourceForm) (interface{}, e.Error) {
	if c.OrgId == "" || c.ProjectId == "" || form.Id == "" {
		return nil, e.New(e.BadRequest, http.StatusBadRequest)
	}

	task, err := services.GetTaskById(c.DB(), form.Id)
	if err != nil && err.Code() != e.TaskNotExists {
		return nil, e.New(err.Code(), err, http.StatusNotFound)
	} else if err != nil {
		c.Logger().Errorf("error get env, err %s", err)
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	}

	query := c.DB().Model(models.Resource{}).Where("org_id = ? AND project_id = ? AND env_id = ? AND task_id = ?",
		c.OrgId, c.ProjectId, task.EnvId, task.Id)

	if form.HasKey("q") {
		// 支持对 provider / type / name 进行模糊查询
		query = query.Where("provider LIKE ? OR type LIKE ? OR name LIKE ?",
			fmt.Sprintf("%%%s%%", form.Q),
			fmt.Sprintf("%%%s%%", form.Q),
			fmt.Sprintf("%%%s%%", form.Q))
	}

	if form.SortField() == "" {
		query = query.Order("provider, type, name")
	}

	rs := make([]models.Resource, 0)
	p := page.New(form.CurrentPage(), form.PageSize(), query)
	if err := p.Scan(&rs); err != nil {
		return nil, e.New(e.DBError, err)
	}

	for i := range rs {
		rs[i].Provider = path.Base(rs[i].Provider)
		// attrs 暂时不需要返回
		rs[i].Attrs = nil
	}
	return &page.PageResp{
		Total:    p.MustTotal(),
		PageSize: p.Size,
		List:     rs,
	}, nil
}
