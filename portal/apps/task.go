// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package apps

import (
	"bufio"
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/libs/page"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
	"cloudiac/utils"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-contrib/sse"
)

// SearchTask 任务查询
func SearchTask(c *ctx.ServiceContext, form *forms.SearchTaskForm) (interface{}, e.Error) {
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
func TaskDetail(c *ctx.ServiceContext, form forms.DetailTaskForm) (*taskDetailResp, e.Error) {
	orgIds, er := services.GetOrgIdsByUser(c.DB(), c.UserId)
	if er != nil {
		c.Logger().Errorf("error get task id by user, err %s", er)
		return nil, e.New(e.DBError, er)
	}
	if !c.OrgId.InArray(orgIds...) && !c.IsSuperAdmin {
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
		c.Logger().Errorf("get task by id err %s", err)
		return nil, e.New(e.DBError, err)
	}
	sort.Sort(task.Variables)
	user, err = services.GetUserByIdRaw(c.DB(), task.CreatorId)
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
	if strings.Contains(o.RepoAddr, `token:`) {
		o.RepoAddr, err = replaceVcsToken(o.RepoAddr)
		if err != nil {
			return nil, err
		}
	}
	return &o, nil
}

func replaceVcsToken(old string) (string, e.Error) {
	u, err := url.Parse(old)
	if err != nil {
		return "", e.New(e.URLParseError, err)
	}
	u.User = url.User("")
	return u.Redacted(), nil
}

// LastTask 最新任务信息
func LastTask(c *ctx.ServiceContext, form *forms.LastTaskForm) (*taskDetailResp, e.Error) {
	if c.OrgId == "" || c.ProjectId == "" {
		return nil, e.New(e.BadRequest, http.StatusBadRequest)
	}
	query := c.DB().Where("org_id = ? AND project_id = ?", c.OrgId, c.ProjectId)
	env, err := services.GetEnvById(query, form.Id)
	if err != nil && err.Code() == e.EnvNotExists {
		return nil, e.New(err.Code(), err, http.StatusNotFound)
	} else if err != nil {
		c.Logger().Errorf("get env by id err %s", err)
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
		c.Logger().Errorf("get task by id err %s", err)
		return nil, e.New(e.DBError, err)
	}
	user, err := services.GetUserByIdRaw(c.DB(), task.CreatorId)
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
func ApproveTask(c *ctx.ServiceContext, form *forms.ApproveTaskForm) (interface{}, e.Error) {
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

func FollowTaskLog(c *ctx.GinRequest, form forms.TaskLogForm) e.Error {
	logger := c.Logger().WithField("func", "FollowTaskLog").WithField("taskId", form.Id)
	sc := c.Service()
	rCtx := c.Context.Request.Context()

	query := services.QueryWithProjectId(services.QueryWithOrgId(sc.DB(), sc.OrgId), sc.ProjectId)
	var tasker models.Tasker
	tasker, er := services.GetTask(query, form.Id)
	if er != nil {
		if sc.IsSuperAdmin {
			tasker, er = services.GetScanTaskById(sc.DB(), form.Id)
		}
		if er != nil {
			logger.Errorf("get task: %v", er)
			if er.Code() == e.TaskNotExists {
				return e.New(er.Code(), http.StatusNotFound)
			}
			return er
		}
	}

	pr, pw := io.Pipe()
	go func() {
		if form.StepId != "" {
			if err := services.FetchTaskStepLog(rCtx, tasker, pw, form.StepId); err != nil {
				logger.Errorf("fetch task step log: %v", err)
			}
		} else {
			if err := services.FetchTaskLog(rCtx, tasker, form.StepType, pw); err != nil {
				logger.Errorf("fetch task log: %v", err)
			}
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
func TaskOutput(c *ctx.ServiceContext, form forms.DetailTaskForm) (interface{}, e.Error) {
	orgIds, er := services.GetOrgIdsByUser(c.DB(), c.UserId)
	if er != nil {
		c.Logger().Errorf("error get task id by user, err %s", er)
		return nil, e.New(e.DBError, er)
	}
	if !c.OrgId.InArray(orgIds...) && !c.IsSuperAdmin {
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
		c.Logger().Errorf("get task by id err %s", err)
		return nil, e.New(e.DBError, err)
	}

	return task.Result.Outputs, nil
}

// SearchTaskResources 查询环境资源列表
func SearchTaskResources(c *ctx.ServiceContext, form *forms.SearchTaskResourceForm) (interface{}, e.Error) {
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

	query := c.DB().Table("iac_resource as r").Where("r.org_id = ? AND r.project_id = ? AND r.env_id = ? AND r.task_id = ?",
		c.OrgId, c.ProjectId, task.EnvId, task.Id)
	query = query.Joins("left join iac_resource_drift as rd on rd.res_id = r.id").LazySelectAppend("r.*, !ISNULL(rd.drift_detail) as is_drift")
	if form.HasKey("q") {
		q := fmt.Sprintf("%%%s%%", form.Q)
		// 支持对 provider / type / name 进行模糊查询
		query = query.Where("r.provider LIKE ? OR r.type LIKE ? OR r.name LIKE ?", q, q, q)
	}

	if form.SortField() == "" {
		query = query.Order("r.provider, r.type, r.name")
	}

	rs := make([]services.Resource, 0)
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

type TaskStepDetail struct {
	Id      models.Id    `json:"id"`
	Index   int          `json:"index"`
	Name    string       `json:"name"`
	TaskId  models.Id    `json:"taskId"`
	Status  string       `json:"status"`
	Message string       `json:"message"`
	StartAt *models.Time `json:"startAt"`
	EndAt   *models.Time `json:"endAt"`
	Type    string       `json:"type"`
}

func SearchTaskSteps(c *ctx.ServiceContext, form *forms.DetailTaskStepForm) (interface{}, e.Error) {
	query := services.QueryTaskStepsById(c.DB(), form.TaskId)
	details := make([]*TaskStepDetail, 0)
	if err := query.Scan(&details); err != nil {
		return nil, e.New(e.DBError, err)
	}
	return details, nil

}

func GetTaskStep(c *ctx.ServiceContext, form *forms.GetTaskStepLogForm) (interface{}, e.Error) {
	content, err := services.GetTaskStepLogById(c.DB(), form.StepId)
	if err != nil {
		return nil, err
	}
	return string(content), nil
}

// SearchTaskResourcesGraph 查询环境资源列表
func SearchTaskResourcesGraph(c *ctx.ServiceContext, form *forms.SearchTaskResourceGraphForm) (interface{}, e.Error) {
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

	rs, err := services.GetTaskResourceToTaskId(c.DB(), task)
	if err != nil {
		return nil, err
	}
	for i := range rs {
		rs[i].Provider = path.Base(rs[i].Provider)
		// attrs 暂时不需要返回
		rs[i].Attrs = nil
	}
	return GetResourcesGraph(rs, form.Dimension), nil
}

func GetResourcesGraph(rs []services.Resource, dimension string) interface{} {
	switch dimension {
	case consts.GraphDimensionModule:
		return GetResourcesGraphModule(rs)
	case consts.GraphDimensionProvider:
		return GetResourcesGraphProvider(rs)
	case consts.GraphDimensionType:
		return GetResourcesGraphType(rs)
	default:
		return nil
	}
}

type ResourcesGraphModule struct {
	NodeId        string                  `json:"nodeId" form:"nodeId" `
	IsRoot        bool                    `json:"isRoot" form:"isRoot" `
	NodeName      string                  `json:"nodeName" form:"nodeName" `
	Children      []*ResourcesGraphModule `json:"children" form:"children" `
	ResourcesList []ResourceInfo          `json:"resourcesList" form:"resourcesList" `
}

type ResourceInfo struct {
	ResourceId   interface{} `json:"resourceId" form:"resourceId" `
	ResourceName string      `json:"resourceName" form:"resourceName" `
	NodeName     string      `json:"nodeName" form:"nodeName" `
	IsDrift      bool        `json:"isDrift"`
}

func GetResourcesGraphModule(resources []services.Resource) interface{} {
	// 构建根节点
	rootModule := "rootModule"
	rgm := &ResourcesGraphModule{
		IsRoot:   true,
		NodeName: rootModule,
		NodeId:   rootModule,
		Children: make([]*ResourcesGraphModule, 0),
	}

	// 存储当前节点与父级节点的关系 父级节点id与子节点关系 {parentNodeId: [nodeId, nodeId]}
	parentChildNode := make(map[string][]string)

	// 资源列表 {nodeId: [resource1,resource2]}
	resourceAttr := make(map[string][]ResourceInfo)

	//查询nodeId对应的nodeName {nodeId: nodeName}
	nodeNameAttr := make(map[string]string)

	/*
		示例： slb.data.alicloud_slbs.this
		{
			nodeId: rootModule
			nodeName: rootModule
			children: [
				{
					"nodeId" :"slb"
					"nodeName" : "slb"
					children : [
						{
							"nodeId" :"slb.data"
							"nodeName" : "data"
							children : [
								{
									"nodeId" :"slb.data.alicloud_slbs"
									"nodeName" : "alicloud_slbs"
									children : []
								}
							]
						}
					]
				}

			]
		}

	*/
	for _, resource := range resources {
		// 将module替替换为空
		address := strings.Replace(resource.Address, "module.", "", -1)
		addrs := strings.Split(address, ".")
		for index, addr := range addrs {
			var (
				parentNodeId string
				nodeId       = strings.Join(addrs[:index+1], ".")
			)
			nodeNameAttr[nodeId] = addr
			if index == 0 {
				parentNodeId = rootModule
			} else {
				parentNodeId = strings.Join(addrs[:index], ".")
			}

			// 处理最末级节点，将末级节点定义为资源
			if index == len(addrs)-1 {
				res := ResourceInfo{
					ResourceId:   resource.Id.String(),
					ResourceName: resource.Name,
					NodeName:     addr,
				}

				if res.ResourceName == "" {
					res.ResourceName = addr
				}

				if resource.DriftDetail != "" {
					res.IsDrift = true
				}

				if _, ok := resourceAttr[parentNodeId]; !ok {
					resourceAttr[parentNodeId] = []ResourceInfo{res}
					continue
				}
				resourceAttr[parentNodeId] = append(resourceAttr[parentNodeId], res)
				continue
			}

			// 构造数据结构
			if _, ok := parentChildNode[parentNodeId]; !ok {
				parentChildNode[parentNodeId] = []string{nodeId}
				continue
			}
			parentChildNode[parentNodeId] = append(parentChildNode[parentNodeId], nodeId)

		}
	}

	// 根据address构造的叶子节点列表有可能重复，这里进行去重
	/*
		a.b.c
		a.b.d
		输出：a: [b,b]
		去重: a: [b]
	*/

	for k, v := range parentChildNode {
		// 去重
		parentChildNode[k] = utils.RemoveDuplicateElement(v)
	}

	// 构造二级节点
	for _, nodeId := range parentChildNode[rgm.NodeId] {
		resourcesGraphModule := &ResourcesGraphModule{
			NodeId:   nodeId,
			IsRoot:   false,
			NodeName: nodeNameAttr[nodeId],
			Children: make([]*ResourcesGraphModule, 0),
		}
		rgm.Children = append(rgm.Children, resourcesGraphModule)
		// 二级节点有可能是末级节点，需要把资源列表放进去
		if _, ok := resourceAttr[nodeId]; ok {
			resourcesGraphModule.ResourcesList = resourceAttr[nodeId]
		}
	}

	getTree(rgm.Children, parentChildNode, resourceAttr, nodeNameAttr)

	return rgm
}

func getTree(parents []*ResourcesGraphModule, parentChildNode map[string][]string,
	resourceAttr map[string][]ResourceInfo, nodeNameAttr map[string]string) {

	for _, parent := range parents {
		for parentId, childIds := range parentChildNode {
			if parentId == parent.NodeId {
				for _, v := range childIds {
					rgm := &ResourcesGraphModule{
						NodeId:   v,
						IsRoot:   false,
						NodeName: nodeNameAttr[v],
						Children: make([]*ResourcesGraphModule, 0),
					}
					if _, ok := resourceAttr[v]; ok {
						rgm.ResourcesList = resourceAttr[v]
					}
					parent.Children = append(parent.Children, rgm)
				}
				break
			}
		}
		// 递归处理叶子节点
		getTree(parent.Children, parentChildNode, resourceAttr, nodeNameAttr)
	}
}

type ProviderTypeResource struct {
	Id      models.Id `json:"id" ` //ID
	Name    string    `json:"name"`
	IsDrift bool      `json:"isDrift"`
}

type ResourcesGraphProvider struct {
	NodeName string                 `json:"nodeName" form:"nodeName" `
	List     []ProviderTypeResource `json:"list" form:"list" `
}

func GetResourcesGraphProvider(rs []services.Resource) interface{} {
	rgt := make([]ResourcesGraphProvider, 0)
	rgtAttr := make(map[string][]ProviderTypeResource)
	for _, v := range rs {
		ptr := ProviderTypeResource{
			Id:   v.Id,
			Name: v.Name,
		}
		if v.DriftDetail != "" {
			ptr.IsDrift = true
		}
		if _, ok := rgtAttr[v.Provider]; !ok {
			rgtAttr[v.Provider] = []ProviderTypeResource{ptr}
			continue
		}
		rgtAttr[v.Provider] = append(rgtAttr[v.Provider], ptr)
	}

	for k, v := range rgtAttr {
		rgt = append(rgt, ResourcesGraphProvider{
			NodeName: k,
			List:     v,
		})
	}

	return rgt
}

type ResourcesGraphType struct {
	NodeName string                 `json:"nodeName" form:"nodeName" `
	List     []ProviderTypeResource `json:"list"`
}

func GetResourcesGraphType(rs []services.Resource) interface{} {
	rgt := make([]ResourcesGraphType, 0)
	rgtAttr := make(map[string][]ProviderTypeResource)
	for _, v := range rs {
		ptr := ProviderTypeResource{
			Id:   v.Id,
			Name: v.Name,
		}
		if v.DriftDetail != "" {
			ptr.IsDrift = true
		}
		if _, ok := rgtAttr[v.Type]; !ok {
			rgtAttr[v.Type] = []ProviderTypeResource{ptr}
			continue
		}
		rgtAttr[v.Type] = append(rgtAttr[v.Type], ptr)
	}
	for k, v := range rgtAttr {
		rgt = append(rgt, ResourcesGraphType{
			NodeName: k,
			List:     v,
		})
	}

	return rgt
}
