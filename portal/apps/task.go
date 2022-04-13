// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package apps

import (
	"bufio"
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/libs/page"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/models/resps"
	"cloudiac/portal/services"
	"cloudiac/utils"
	"cloudiac/utils/logs"
	"context"
	"errors"
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
		query = query.Where("env_id = ? AND is_drift_task != 1 OR  (is_drift_task = 1 AND applied = ?)",
			form.EnvId, true)
	}
	//根据任务类型查询
	if form.TaskType != "" {
		query = query.Where("iac_task.type = ?", form.TaskType)
	}
	//根据触发类型查询
	if form.Source != "" {
		query = query.Where("iac_task.source = ?", form.Source)
	}
	//根据执行人名称或邮箱查询
	if form.User != "" {
		users := "%" + form.User + "%"
		query = query.Where("u.name like ?  or u.email LIKE ?", users, users)
	}
	// 默认按创建时间逆序排序
	if form.SortField() == "" {
		query = query.Order("created_at DESC")
	}

	p := page.New(form.CurrentPage(), form.PageSize(), query)
	details := make([]*resps.TaskDetailResp, 0)
	if err := p.Scan(&details); err != nil {
		return nil, e.New(e.DBError, err)
	}

	for _, env := range details {
		// 隐藏敏感字段
		env.HideSensitiveVariable()
	}

	return page.PageResp{
		Total:    p.MustTotal(),
		PageSize: p.Size,
		List:     details,
	}, nil
}

// TaskDetail 任务信息详情
func TaskDetail(c *ctx.ServiceContext, form forms.DetailTaskForm) (*resps.TaskDetailResp, e.Error) {
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
		c.Logger().Errorf("error get task by id, err %s", err)
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
	var o = resps.TaskDetailResp{
		Task:    *task,
		Creator: user.Name,
	}
	// 清除url token
	o.RepoAddr, err = replaceVcsToken(o.RepoAddr)
	if err != nil {
		return nil, err
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
func LastTask(c *ctx.ServiceContext, form *forms.LastTaskForm) (*resps.TaskDetailResp, e.Error) {
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
	var t = resps.TaskDetailResp{
		Task:    *task,
		Creator: user.Name,
	}

	return &t, nil
}

// ApproveTask 审批执行计划
func ApproveTask(c *ctx.ServiceContext, form *forms.ApproveTaskForm) (interface{}, e.Error) { //nolint:cyclop
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
		return nil, e.New(e.TaskApproveNotPending, http.StatusConflict)
	}

	if task.Aborting {
		return nil, e.New(e.TaskAborting, http.StatusConflict)
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

func getTask(sc *ctx.ServiceContext, id models.Id) (models.Tasker, e.Error) {
	query := services.QueryWithProjectId(services.QueryWithOrgId(sc.DB(), sc.OrgId), sc.ProjectId)

	var (
		tasker models.Tasker
		er     e.Error
	)
	tasker, er = services.GetTask(query, id)
	if er != nil {
		if sc.IsSuperAdmin {
			tasker, er = services.GetScanTaskById(sc.DB(), id)
		}
		if er != nil {
			if er.Code() == e.TaskNotExists {
				return nil, e.New(er.Code(), http.StatusNotFound)
			}
			return nil, er
		}
	}

	return tasker, nil
}

func startTaskLog(rCtx context.Context, tasker models.Tasker, pw *io.PipeWriter, form forms.TaskLogForm, logger logs.Logger) {
	if form.StepId != "" {
		if err := services.FetchTaskStepLog(rCtx, tasker, pw, form.StepId); err != nil {
			logger.Errorf("fetch task step log: %v", err)
		}
	} else {
		if err := services.FetchTaskLog(rCtx, tasker, form.StepType, pw); err != nil {
			logger.Errorf("fetch task log: %v", err)
		}
	}
}

// TODO tasker 的逻辑有疑问
func FollowTaskLog(c *ctx.GinRequest, form forms.TaskLogForm) e.Error {
	logger := c.Logger().WithField("func", "FollowTaskLog").WithField("taskId", form.Id)
	sc := c.Service()
	rCtx := c.Context.Request.Context()

	tasker, er := getTask(sc, form.Id)
	if er != nil {
		return er
	}

	pr, pw := io.Pipe()
	go startTaskLog(rCtx, tasker, pw, form, logger)

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

	if err := scanner.Err(); err != nil && !errors.Is(err, io.EOF) {
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
		c.Logger().Errorf("error get task by id, err %s", err)
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

func SearchTaskSteps(c *ctx.ServiceContext, form *forms.DetailTaskStepForm) (interface{}, e.Error) {
	query := services.QueryTaskStepsById(c.DB(), form.TaskId)
	details := make([]*resps.TaskStepDetail, 0)
	if err := query.Scan(&details); err != nil {
		return nil, e.New(e.DBError, err)
	}
	return details, nil

}

func GetTaskStepLog(c *ctx.ServiceContext, form *forms.GetTaskStepLogForm) (interface{}, e.Error) {
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
		// 不需要返回 attrs, 避免返回的数据过大;
		if form.Dimension != consts.GraphDimensionModule {
			rs[i].Attrs = nil
		}
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

// GetResShowName
// 建立规则库，通过各种规则确定资源的主要字段或者展示模板
//		规则示例1: 如果资源的属性中有 public_ip 字段，则展示 public_ip;
//    	规则示例2: 如果资源的属性中有 name 字段，则展示 name;
//    	规则示例3: 如果资源的属性中有 tag 字段，则展示 name(tag1,tag2);
// 不匹配规则库时展示: resource address(id), 如: "module1.alicloud_instance.web(i-xxxxxxx)";
func GetResShowName(attrs map[string]interface{}, addr string) string {
	get := func(key string) (string, bool) {
		// 如果 val 为空字符则视为无值
		if val, ok := attrs[key]; ok {
			switch OriginalValue := val.(type) {
			case map[string]string:
				var expectedFormat = make([]string, 0) // expect format: "k1=v1,k2=v2,k3=v3..."
				for k, v := range OriginalValue {
					expectedFormat = append(expectedFormat, fmt.Sprintf("%s=%s", k, v))
				}
				if len(expectedFormat) == 0 {
					return "", false
				}
				return strings.Join(expectedFormat, ","), true
			case []string:
				expectFormat := strings.Join(OriginalValue, ",") // expect format: "v1,v2,v3..."
				if len(expectFormat) == 0 {
					return "", false
				}
				return expectFormat, true
			case nil:
				return "", false
			default:
				str := fmt.Sprintf("%v", OriginalValue)
				if len(str) == 0 {
					return str, false
				}
				return str, true
			}
		}
		return "", false
	}

	if attrs != nil {
		if publicIP, ok := get("public_ip"); ok {
			return publicIP
		}
		if name, ok := get("name"); ok {
			if tags, ok := get("tags"); ok {
				return fmt.Sprintf("%s(%s)", name, tags)
			}
			return name
		}
	}
	outRuleName, ok := get("id")
	if ok {
		return fmt.Sprintf("%s(%s)", addr, outRuleName)
	}
	return addr
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

func genNodesFromResource(resource services.Resource, parentChildNode map[string][]string, resourceAttr map[string][]ResourceInfo, nodeNameAttr map[string]string) {
	// 将module替替换为空
	address := strings.Replace(resource.Address, "module.", "", -1)
	addrs := strings.Split(address, ".")
	if len(addrs) == 0 {
		return
	}

	// first node
	rootModule := "rootModule"
	rootNodeId := strings.Join(addrs[:1], ".")
	nodeNameAttr[rootNodeId] = addrs[0]
	if _, ok := parentChildNode[rootModule]; !ok {
		parentChildNode[rootModule] = []string{rootNodeId}
	} else {
		parentChildNode[rootModule] = append(parentChildNode[rootModule], rootNodeId)
	}

	// middle nodes
	for index := 1; index < len(addrs)-1; index++ {
		var (
			parentNodeId string
			nodeId       = strings.Join(addrs[:index+1], ".")
		)
		nodeNameAttr[nodeId] = addrs[index]
		parentNodeId = strings.Join(addrs[:index], ".")

		// 构造数据结构
		if _, ok := parentChildNode[parentNodeId]; !ok {
			parentChildNode[parentNodeId] = []string{nodeId}
			continue
		}
		parentChildNode[parentNodeId] = append(parentChildNode[parentNodeId], nodeId)
	}

	// last node 处理最末级节点，将末级节点定义为资源
	lastAddr := addrs[len(addrs)-1]
	lastNodeId := strings.Join(addrs[:], ".")
	nodeNameAttr[lastNodeId] = lastAddr
	lastParentNodeId := strings.Join(addrs[:len(addrs)-1], ".")

	res := ResourceInfo{
		ResourceId:   resource.Id.String(),
		ResourceName: GetResShowName(resource.Attrs, resource.Address),
		NodeName:     lastAddr,
	}

	if res.ResourceName == "" {
		res.ResourceName = lastAddr
	}

	if resource.DriftDetail != "" {
		res.IsDrift = true
	}

	if _, ok := resourceAttr[lastParentNodeId]; !ok {
		resourceAttr[lastParentNodeId] = []ResourceInfo{res}
	} else {
		resourceAttr[lastParentNodeId] = append(resourceAttr[lastParentNodeId], res)
	}
}

func genNodesFromAllResources(resources []services.Resource) (map[string][]string, map[string][]ResourceInfo, map[string]string) {
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
		genNodesFromResource(resource, parentChildNode, resourceAttr, nodeNameAttr)
	}

	return parentChildNode, resourceAttr, nodeNameAttr
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
	parentChildNode, resourceAttr, nodeNameAttr := genNodesFromAllResources(resources)

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

func AbortTask(c *ctx.ServiceContext, form *forms.AbortTaskForm) (interface{}, e.Error) {
	er := c.DB().Transaction(func(tx *db.Session) error {
		task, er := services.GetTaskById(tx, form.TaskId)
		if er != nil {
			return er
		}

		if task.Aborting {
			return e.New(e.TaskAborting)
		}

		step, er := services.GetTaskStep(tx, task.Id, task.CurrStep)
		if er != nil {
			return er
		}

		if task.Status == models.TaskPending {
			task.Status = models.TaskAborted
			if _, err := models.UpdateModel(tx, task); err != nil {
				return e.AutoNew(err, e.DBError)
			}
		} else if step.Status == models.TaskStepApproving {
			task.Aborting = true
			if _, err := models.UpdateModel(tx, task); err != nil {
				return e.AutoNew(err, e.DBError)
			}
			// 步骤在待审批状态时直接将状态改为 aborted 并同步修改任务状态
			if er := services.ChangeTaskStep2Aborted(tx, task.Id, step.Index); er != nil {
				return er
			}
		} else if task.Started() && !task.Exited() {
			if err := services.CheckRunnerTaskCanAbort(*task); err != nil {
				return e.New(e.TaskCannotAbort, err)
			}

			task.Aborting = true
			if _, err := models.UpdateModel(tx, task); err != nil {
				return e.AutoNew(err, e.DBError)
			}

			// 任务在执行状态时发送指令中断 runner 的任务执行，然后 runner 会上报步骤被中止
			go utils.RecoverdCall(func() {
				goAbortRunnerTask(c.Logger(), *task)
			})
		} else {
			return e.New(e.TaskCannotAbort,
				fmt.Errorf("task status is '%s'", task.Status), http.StatusConflict)
		}
		return nil
	})

	if er != nil {
		return nil, e.AutoNew(er, e.InternalError)
	}
	return nil, nil
}

func goAbortRunnerTask(logger logs.Logger, task models.Task) {
	logger = logger.WithField("action", "goAbortRunnerTask")
	if er := services.AbortRunnerTask(task); er != nil {
		logger.Errorf("abort task error: %v", er)
		task.Aborting = false
		if _, err := models.UpdateAttr(db.Get(), &models.Task{},
			models.Attrs{"aborting": false}, "id = ?", task.Id); err != nil {
			logger.Errorf("update task aborting error: %v", err)
		}
	}
}
