package services

import (
	"bufio"
	"cloudiac/configs"
	"cloudiac/consts"
	"cloudiac/consts/e"
	"cloudiac/libs/db"
	"cloudiac/models"
	"cloudiac/models/forms"
	"cloudiac/utils"
	"cloudiac/utils/logs"
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	"net/http"
	"os"
	"strings"
	"time"
)

func CreateTask(tx *db.Session, task models.Task) (*models.Task, e.Error) {
	if err := models.Create(tx, &task); err != nil {
		if e.IsDuplicate(err) {
			return nil, e.New(e.TaskAlreadyExists, err)
		}
		return nil, e.New(e.DBError, err)
	}

	return &task, nil
}

func UpdateTask(tx *db.Session, id uint, attrs models.Attrs) (org *models.Task, re e.Error) {
	org = &models.Task{}
	if _, err := models.UpdateAttr(tx.Where("id = ?", id), &models.Task{}, attrs); err != nil {
		return nil, e.New(e.DBError, fmt.Errorf("update task error: %v", err))
	}
	if err := tx.Where("id = ?", id).First(org); err != nil {
		return nil, e.New(e.DBError, fmt.Errorf("query task error: %v", err))
	}
	return
}

func GetTaskById(tx *db.Session, id uint) (*models.Task, e.Error) {
	o := models.Task{}
	if err := tx.Where("id = ?", id).First(&o); err != nil {
		if e.IsRecordNotFound(err) {
			return nil, e.New(e.TaskNotExists, err)
		}
		return nil, e.New(e.DBError, err)
	}
	return &o, nil
}

func QueryTask(tx *db.Session, status, q string, tplId uint) *db.Session {
	query := tx.Table(fmt.Sprintf("%s as task", models.Task{}.TableName())).
		Where("template_id = ?", tplId).
		Joins(fmt.Sprintf("left join %s as tpl on tpl.id = task.template_id", models.Template{}.TableName())).
		LazySelectAppend("task.*, tpl.repo_branch")
	if status != "" {
		query = query.Where("task.status = ?", status)
	}
	if q != "" {
		qs := "%" + q + "%"
		query = query.Where("task.name LIKE ? OR task.description LIKE ?", qs, qs)
	}

	return query.Order("task.created_at DESC")

}

func TaskDetail(tx *db.Session, taskId uint) *db.Session {
	return tx.Table(models.Task{}.TableName()).Select(fmt.Sprintf("%s.*, tpl.*", models.Task{}.TableName())).
		Joins(fmt.Sprintf("left join %s as tpl on tpl.id = %s.template_id", models.Template{}.TableName(), models.Task{}.TableName())).
		Where(fmt.Sprintf("%s.id = %d", models.Task{}.TableName(), taskId))
}

func LastTask(tx *db.Session, tplId uint) *db.Session {
	return tx.Table(models.Task{}.TableName()).Where("template_id = ?", tplId)
}

type LastTaskInfo struct {
	Status    string    `json:"status"`
	Guid      string    `json:"taskGuid"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func GetTaskByTplId(tx *db.Session, tplId uint) (*LastTaskInfo, e.Error) {
	lastTaskInfo := LastTaskInfo{}
	err := tx.Table(models.Task{}.TableName()).
		Select("status, guid, updated_at").
		Where("template_id = ?", tplId).
		//Where("status in (?)",statusList).
		Find(&lastTaskInfo)
	if err != nil {
		return nil, e.New(e.DBError, err)
	}
	return &lastTaskInfo, nil
}

//var (
//	taskTicker *time.Ticker = time.NewTicker(time.Duration(configs.Get().Task.TimeTicker) * time.Second)
//	runnerAddr string       = configs.Get().Task.Addr
//	//runnerAddr string       = ""
//)

func runningTaskEnvParam(tpl *models.Template, runnerId string, orgId uint) interface{} {
	vars := make([]forms.Var, 0)
	param := make(map[string]interface{})
	varsByte, _ := tpl.Vars.MarshalJSON()
	if !tpl.Vars.IsNull() {
		json.Unmarshal(varsByte, &vars)
	}
	vars = append(vars, resourceEnvParam(runnerId, orgId)...)
	for _, v := range vars {
		if v.Type == consts.Terraform {
			v.Key = fmt.Sprintf("%s%s", consts.TerraformVar, v.Key)
		}
		if *v.IsSecret {
			param[v.Key] = utils.AesDecrypt(v.Value)
		} else {
			param[v.Key] = v.Value
		}
	}

	return param
}

func resourceEnvParam(runnerId string, orgId uint) []forms.Var {
	vars := make([]forms.Var, 0)
	ra := []models.ResourceAccount{}
	//org,_:=getorg
	if err := db.Get().Debug().Joins(fmt.Sprintf("left join %s as crm on %s.id = crm.resource_account_id",
		models.CtResourceMap{}.TableName(), models.ResourceAccount{}.TableName())).
		Where("crm.ct_service_id = ?", runnerId).
		Where(fmt.Sprintf("%s.status = '%s'", models.ResourceAccount{}.TableName(), consts.ResourceAccountEnable)).
		Where(fmt.Sprintf("%s.org_id = %d", models.ResourceAccount{}.TableName(), orgId)).
		Find(&ra); err != nil {
		logs.Get().Errorf("ResourceAccount db err %v: ", err)
		return nil
	}

	for _, raInfo := range ra {
		varsByte, _ := raInfo.Params.MarshalJSON()
		if !raInfo.Params.IsNull() {
			v := make([]forms.Var, 0)
			json.Unmarshal(varsByte, &v)
			vars = append(vars, v...)
		}
	}

	return vars
}

func getBackendInfo(backendInfo models.JSON, containerId string) []byte {
	attr := models.Attrs{}
	json.Unmarshal(backendInfo, &attr)
	attr["container_id"] = containerId
	b, _ := json.Marshal(attr)
	return b
}

func updateBackendInfo(backendInfo models.JSON, offset int) []byte {
	attr := models.Attrs{}
	json.Unmarshal(backendInfo, &attr)
	attr["log_offset"] = attr["log_offset"].(float64) + float64(offset)
	b, _ := json.Marshal(attr)
	return b
}

func writeTaskLog(contentList []string, logPath string, offset float64) error {
	path := fmt.Sprintf("%s/%s", logPath, consts.TaskLogName)
	var (
		file *os.File
		err  error
	)
	isExists, _ := utils.PathExists(path)
	if !isExists {
		file, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0666)
	} else {
		file, err = os.OpenFile(path, os.O_WRONLY|os.O_APPEND, 0666)
	}

	if err != nil {
		return err
	}
	defer file.Close()
	write := bufio.NewWriter(file)
	for _, content := range contentList {
		write.WriteString(content)
	}
	write.Flush()

	return nil
}

func RunTaskToRunning(task *models.Task, dbsess *db.Session, orgGuid string) {
	logger := logs.Get().WithField("action", "RunTaskToRunning")
	conf := configs.Get()

	tx := dbsess.Begin()
	tpl := &models.Template{}
	var taskStatus string
	for {
		count, _ := dbsess.Table(models.Task{}.TableName()).
			Where("ct_service_id = ?", task.CtServiceId).
			Where("status = ?", consts.TaskRunning).
			Count()

		if int(count) > utils.GetRunnerMax() {
			tx.Commit()
			logger.Infof("runner concurrent num gt %d", count)
			time.Sleep(time.Second * 5)
			continue
		}

		//获取模板参数
		if err := tx.
			Where("id = ?", task.TemplateId).
			First(tpl); err != nil && err != gorm.ErrRecordNotFound {
			tx.Commit()
			logger.Errorf("tpl db err: %v", err)
			time.Sleep(time.Second * 5)
			continue
		}
		taskBackend := make(map[string]interface{}, 0)
		json.Unmarshal(task.BackendInfo, &taskBackend)

		//向runner下发task
		runnerAddr := taskBackend["backend_url"]
		addr := fmt.Sprintf("%s%s", runnerAddr, "/task/run")
		//有状态云模版，以模版ID为路径，无状态云模版，以模版ID + 作业ID 为路径
		var stateKey string
		if tpl.SaveState {
			stateKey = fmt.Sprintf("%s/%s.tfstate", orgGuid, tpl.Guid)
		} else {
			stateKey = fmt.Sprintf("%s/%s/%s.tfstate", orgGuid, tpl.Guid, task.Guid)
		}
		repoList := strings.Split(tpl.RepoAddr, "//")
		repoAddr := tpl.RepoAddr
		if len(repoList) == 2 {
			repoAddr = fmt.Sprintf("%s//%s:%s@%s", repoList[0], conf.Gitlab.Username, conf.Gitlab.Token, repoList[1])
		}

		data := map[string]interface{}{
			"repo":          repoAddr,
			"template_uuid": tpl.Guid,
			"task_id":       task.Guid,
			//"task_id":       strconv.Itoa(int(task.Id)),
			"state_store": map[string]interface{}{
				"save_state":            tpl.SaveState,
				"backend":               "consul",
				"scheme":                "http",
				"state_key":             stateKey,
				"state_backend_address": conf.Consul.Address,
			},
			"env":     runningTaskEnvParam(tpl, task.CtServiceId, tpl.OrgId),
			"varfile": tpl.Varfile,
			"mode":    task.TaskType,
			"extra":   tpl.Extra,
		}
		header := &http.Header{}
		header.Set("Content-Type", "application/json")
		logger.Tracef("post data: %#v", data)
		//fmt.Printf("post data: %+v /n ", data)

		respData, err := utils.HttpService(addr, "POST", header, data, 20, 5)

		if err != nil {
			logger.Errorf("request failed: %v", err)
		}
		logger.Debugf("response body: %s", string(respData))

		var (
			runnerResp struct {
				Id    string `json:"id" form:"id" `
				Code  string `json:"code" form:"code" `
				Error string `json:"err" form:"err" `
			}
			status  string
			startAt = time.Now()
		)

		if err := json.Unmarshal(respData, &runnerResp); err != nil {
			logger.Errorf("unmarshal error: %v, body: %s", err, string(respData))
		}
		//考虑runner挂掉的情况 使用镜像id作为条件
		if runnerResp.Id == "" {
			status = consts.TaskFailed
			logger.Errorf("Code: %s, Message: %s, Id: %s", runnerResp.Code, runnerResp.Error, runnerResp.Id)
		} else {
			status = consts.TaskRunning
		}
		//更新task状态, 同时生成backend_info并更新
		if _, err := tx.
			Table(models.Task{}.TableName()).
			Where("id = ?", task.Id).
			Update(map[string]interface{}{
				"status": status,
				//这里是第一生成backend直接修改即可
				"start_at":     startAt,
				"backend_info": models.JSON(getBackendInfo(task.BackendInfo, runnerResp.Id)),
			}); err != nil {
			if err := tx.Commit(); err != nil {
				tx.Rollback()
			}
		}
		task.Status = status
		task.StartAt = &startAt
		task.BackendInfo = models.JSON(getBackendInfo(task.BackendInfo, runnerResp.Id))
		if err := tx.Commit(); err != nil {
			tx.Rollback()
		}
		taskStatus = status
		break
	}

	for taskStatus == consts.TaskRunning {
		taskStatus = RunTaskState(task, tpl, dbsess)
		//不能太频繁调用  这里睡一会
		time.Sleep(time.Second * 5)
	}

}

func RunTaskState(task *models.Task, tpl *models.Template, dbsess *db.Session) string {
	logger := logs.Get().WithField("action", "RunTaskState")
	tx := dbsess.Begin()
	taskBackend := make(map[string]interface{}, 0)
	json.Unmarshal(task.BackendInfo, &taskBackend)
	runnerAddr := taskBackend["backend_url"]
	addr := fmt.Sprintf("%s%s", runnerAddr, "/task/status")

	data := map[string]interface{}{
		"template_uuid": tpl.Guid,
		"task_id":       task.Guid,
		"container_id":  taskBackend["container_id"],
		"offset":        taskBackend["log_offset"],
	}
	header := &http.Header{}
	header.Set("Content-Type", "application/json")
	logger.Tracef("post data: %#v", data)

	respData, err := utils.HttpService(addr, "POST", header, data, 20, 5)
	if err != nil {
		logger.Errorf("request failed: %v", err)
	}
	logger.Tracef("response body: %s", string(respData))

	var (
		runnerResp struct {
			Status          string   `json:"status" form:"status" `
			StatusCode      int      `json:"status_code" form:"status_code" `
			LogContent      []string `json:"log_content" form:"log_content" `
			LogContentLines int      `json:"log_content_lines" form:"log_content_lines" `
			Code            string   `json:"code" form:"code" `
			Error           string   `json:"error" form:"error" `
		}
		status = consts.TaskRunning
	)

	if err := json.Unmarshal(respData, &runnerResp); err != nil {
		logger.Errorf("unmarshal error: %v, body: %s", err, string(respData))
	}

	if err := writeTaskLog(runnerResp.LogContent, taskBackend["log_file"].(string), taskBackend["log_offset"].(float64)); err != nil {
		logger.Errorf("write task log error: %v", err)
	}

	if task.StartAt.Unix()+tpl.Timeout < time.Now().Unix() {
		status = consts.TaskTimeoout
	} else if runnerResp.Status == consts.DockerStatusExited && runnerResp.StatusCode != 0 {
		status = consts.TaskFailed
	} else if runnerResp.Status == consts.DockerStatusExited && runnerResp.StatusCode == 0 {
		status = consts.TaskComplete
	}

	updateM := map[string]interface{}{
		"status":       status,
		"backend_info": updateBackendInfo(task.BackendInfo, runnerResp.LogContentLines),
	}

	if status != consts.TaskRunning {
		updateM["end_at"] = time.Now()
	}

	//更新task状态
	if _, err := tx.
		Table(models.Task{}.TableName()).
		Where("id = ?", task.Id).
		Update(updateM); err != nil {
		if err := tx.Commit(); err != nil {
			tx.Rollback()
			return status
		}
	}
	task.Status = status
	task.BackendInfo = updateBackendInfo(task.BackendInfo, runnerResp.LogContentLines)

	if err := tx.Commit(); err != nil {
		tx.Rollback()
	}

	return status
}

func RunTask() {
	dbsess := db.Get()
	taskList := []models.Task{}
	logger := logs.Get()
	if err := dbsess.Where("status = ? or status = ?", consts.TaskPending, consts.TaskRunning).Find(&taskList); err != nil {
		logger.Errorf("RunTask task db err: %v", err)
		return
	}
	for index, _ := range taskList {
		tpl := models.Template{}
		if err := dbsess.Where("id = ?", taskList[index].TemplateId).First(&tpl); err != nil {
			logger.Errorf("RunTask tpl db err: %v, task_id: %d", err, taskList[index].Id)
			continue
		}
		if taskList[index].Status == consts.TaskPending {
			org := models.Organization{}
			if err := dbsess.Where("id = ?", tpl.OrgId).First(&org); err != nil {
				logger.Errorf("RunTask org db err: %v, task_id: %d, tpl_id: %d", err, taskList[index].Id, tpl.Id)
				continue
			}
			go RunTaskToRunning(&taskList[index], dbsess, org.Guid)
		}
		if taskList[index].Status == consts.TaskRunning {
			go func() {
				taskStatus := consts.TaskRunning
				for taskStatus == consts.TaskRunning {
					taskStatus = RunTaskState(&taskList[index], &tpl, dbsess)
					time.Sleep(time.Second * 5)
				}
			}()
		}
		time.Sleep(time.Second)
	}
}
