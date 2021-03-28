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
	"net/http"
	"os"
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

func QueryTask(query *db.Session) *db.Session {
	return query.Model(&models.Task{})
}

func TaskDetail(tx *db.Session, taskId uint) *db.Session {
	return tx.Table(models.Task{}.TableName()).Select(fmt.Sprintf("%s.*, tpl.*", models.Task{}.TableName())).
		Joins(fmt.Sprintf("left join %s as tpl on tpl.id = %s.template_id", models.Template{}.TableName(), models.Task{}.TableName())).
		Where(fmt.Sprintf("%s.id = %d", models.Task{}.TableName(), taskId))
}

type LastTaskInfo struct {
	Status     string  `json:"status"`
	Guid       string  `json:"taskGuid"`
	UpdatedAt  time.Time `json:"updatedAt"`
}
func GetTaskByTplId(tx *db.Session, tplId uint) (*LastTaskInfo, e.Error) {
	lastTaskInfo := LastTaskInfo{}
	err := tx.Table(models.Task{}.TableName()).Select("status, guid, updated_at").Where("template_id = ?", tplId).Find(&lastTaskInfo)
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

func runningTaskEnvParam(tpl models.Template) interface{} {
	vars := make([]forms.Var, 0)
	param := make(map[string]interface{})
	varsByte, _ := tpl.Vars.MarshalJSON()
	if !tpl.Vars.IsNull() {
		json.Unmarshal(varsByte, &vars)
	}
	for _, v := range vars {
		if v.IsSecret {
			param[v.Key] = utils.AesDecrypt(v.Value)
		} else {
			param[v.Key] = v.Value
		}
	}

	return param
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

func RunTaskToRunning() {
	logger := logs.Get().WithField("action", "RunTaskToRunning")
	dbsess := db.Get().Debug()
	conf := configs.Get()
	var (
		taskTicker *time.Ticker = time.NewTicker(time.Duration(conf.Task.TimeTicker) * time.Second)
		runnerAddr string       = "http://127.0.0.1:19030/api/v1"
		//runnerAddr string       = ""
	)
	for {
		go func() {
			tx := dbsess.Begin()
			task := models.Task{}
			tpl := models.Template{}
			//获取状态为pending的任务 查询时增加行锁
			if err := tx.Set("gorm:query_option", "FOR UPDATE").
				Where("status = ?", consts.TaskPending).
				First(&task); err != nil {
				tx.Commit()
				return
			}
			//获取模板参数
			if err := tx.
				Where("id = ?", task.TemplateId).
				First(&tpl); err != nil {
				tx.Commit()
				return
			}
			//向runner下发task
			addr := fmt.Sprintf("%s%s", runnerAddr, "/task/run")
			//todo 入参
			data := map[string]interface{}{
				"repo":          tpl.RepoAddr,
				"template_uuid": tpl.Guid,
				"task_id":       task.Guid,
				// todo taskGuid 生成
				//"task_id":       strconv.Itoa(int(task.Id)),
				"state_store": map[string]interface{}{
					"save_state":            tpl.SaveState,
					"backend":               "consul",
					"scheme":                "http",
					"state_key":             fmt.Sprintf("%s.tfstate", tpl.Guid),
					"state_backend_address": conf.Consul.Address,
				},
				"env":     runningTaskEnvParam(tpl),
				"varfile": tpl.Varfile,
				"mode":    task.TaskType,
				"extra":   tpl.Extra,
			}
			header := &http.Header{}
			header.Set("Content-Type", "application/json")
			logger.Tracef("post data: %#v", data)
			fmt.Printf("post data: %+v /n ", data)

			respData, err := utils.HttpService(addr, "POST", header, data, 20, 5)
			if err != nil {
				logger.Errorf("request failed: %v", err)
			}
			logger.Tracef("response body: %s", string(respData))

			var (
				runnerResp struct {
					Id    string `json:"id" form:"id" `
					Code  string `json:"code" form:"code" `
					Error string `json:"err" form:"err" `
				}
				status string
			)

			if err := json.Unmarshal(respData, &runnerResp); err != nil {
				logger.Errorf("unmarshal error: %v, body: %s", err, string(respData))
			}

			if runnerResp.Error != "" {
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
					"backend_info": models.JSON(getBackendInfo(task.BackendInfo, runnerResp.Id)),
				}); err != nil {
				if err := tx.Commit(); err != nil {
					tx.Rollback()
				}
			}
			if err := tx.Commit(); err != nil {
				tx.Rollback()
			}
		}()
		//time.Sleep(1000000*time.Second)
		<-taskTicker.C
	}
}

func writeTaskLog(contentList []string, logPath string) error {
	path := fmt.Sprintf("%s/%s", logPath, consts.TaskLogName)

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0666)
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

func RunTaskState() {
	logger := logs.Get().WithField("action", "RunTaskState")
	dbsess := db.Get().Debug()
	taskTicker := time.NewTicker(time.Duration(configs.Get().Task.TimeTicker) * time.Second)
	runnerAddr := "http://127.0.0.1:19030/api/v1"

	for {
		go func() {
			tx := dbsess.Begin()
			// 查询running状态的task 查询时增加行锁
			task := models.Task{}
			tpl := models.Template{}
			taskBackend := make(map[string]interface{}, 0)
			if err := tx.Set("gorm:query_option", "FOR UPDATE").
				Where("status = ?", consts.TaskRunning).
				First(&task); err != nil {
				tx.Commit()
				return
			}

			//获取模板参数
			if err := tx.
				Where("id = ?", task.TemplateId).
				First(&tpl); err != nil {
				tx.Commit()
				return
			}
			json.Unmarshal(task.BackendInfo, &taskBackend)
			addr := fmt.Sprintf("%s%s", runnerAddr, "/task/status")
			fmt.Println(addr)

			data := map[string]interface{}{
				"template_uuid": tpl.Guid,
				"task_id":       task.Guid,
				"container_id":  taskBackend["container_id"],
				"offset":        taskBackend["log_offset"],
			}
			header := &http.Header{}
			header.Set("Content-Type", "application/json")
			logger.Tracef("post data: %#v", data)
			fmt.Printf("post data: %#v /n", data)

			respData, err := utils.HttpService(addr, "POST", header, data, 20, 5)
			if err != nil {
				logger.Errorf("request failed: %v", err)
			}
			logger.Tracef("response body: %s", string(respData))

			var (
				runnerResp struct {
					Status          string   `json:"status" form:"status" `
					StatusCode      int      `json:"statusCode" form:"statusCode" `
					LogContent      []string `json:"log_content" form:"log_content" `
					LogContentLines int      `json:"log_content_lines" form:"log_content_lines" `
					Code            string   `json:"code" form:"code" `
					Error           string   `json:"error" form:"error" `
				}
				status string = consts.TaskRunning
			)

			if err := json.Unmarshal(respData, &runnerResp); err != nil {
				logger.Errorf("unmarshal error: %v, body: %s", err, string(respData))
			}

			if err := writeTaskLog(runnerResp.LogContent, taskBackend["log_file"].(string)); err != nil {
				logger.Errorf("write task log error: %v", err)
			}

			if runnerResp.Status == "exised" && runnerResp.StatusCode != 0 {
				status = consts.TaskFailed
			} else {
				status = consts.TaskComplete
			}
			fmt.Println(status)
			//更新task状态
			if _, err := tx.
				Table(models.Task{}.TableName()).
				Where("id = ?", task.Id).
				Update(map[string]interface{}{
					//"status":       status,
					"backend_info": updateBackendInfo(task.BackendInfo, runnerResp.LogContentLines),
				}); err != nil {
				if err := tx.Commit(); err != nil {
					tx.Rollback()
				}
			}
			if err := tx.Commit(); err != nil {
				tx.Rollback()
			}
			// todo log保存
		}()

		<-taskTicker.C
	}

}
