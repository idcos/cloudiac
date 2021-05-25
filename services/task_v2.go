package services

import (
	"cloudiac/configs"
	"cloudiac/consts/e"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"net/url"
	"runtime/debug"
	"time"

	"cloudiac/consts"
	"cloudiac/libs/db"
	"cloudiac/models"
	"cloudiac/runner"
	"cloudiac/utils"
	"cloudiac/utils/kafka"
	"cloudiac/utils/logs"
)

// StartTask 下发任务并等待结束
func StartTask(dbSess *db.Session, orgGuid string, task models.Task) {
	logger := logs.Get().WithField("action", "StartTask").WithField("taskId", task.Guid)

	var (
		dbTask *models.Task
		err error
	)

	tpl := models.Template{}
	if err = dbSess.Where("id = ?", task.TemplateId).First(&tpl); err != nil {
		if e.IsRecordNotFound(err) {
			task.Status = consts.TaskFailed
			task.StatusDetail = fmt.Errorf("tplId '%d' not found", task.TemplateId).Error()
			if _, err := dbSess.Model(&task).Update(&task); err != nil {
				logger.Errorln(err)
			}
		}
		logger.Errorf("query template '%d' error: %v", task.TemplateId, err)
		return
	}

	//TODO 临时方案，一直循环调用直到任务 Exited, 后续应该改由 taskManager 统一重试
	for {
		AssignTask(dbSess, orgGuid, task, &tpl)

		if dbTask, err = GetTaskById(dbSess, task.Id); err != nil {
			logger.Errorln(err)
		} else if dbTask.Exited() || dbTask.Status == consts.TaskRunning {
			break
		}
		time.Sleep(time.Second * 5)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var taskStatus string
	if dbTask.Status == consts.TaskRunning {
		var err error
		taskStatus, err = WaitTask(ctx, task.Guid, &tpl, dbSess)
		if err != nil {
			logger.Errorf("wait task error: %v", err)
			return
		}
	}

	logger.Infof("task exited, status: %s", taskStatus)

	//todo 作业执行完成之后 日志有可能拿不完
	getTaskLogs(task.TemplateGuid, task.Guid, dbSess)
	if taskStatus  == consts.TaskComplete {
		logPath := task.BackendInfo
		path := map[string]interface{}{}
		json.Unmarshal(logPath, &path)
		if logFile, ok := path["log_file"].(string); ok {
			runnerFilePath := logFile
			tfInfo := GetTFLog(runnerFilePath)
			models.UpdateAttr(dbSess, dbTask, tfInfo)
		}
	}
}

// AssignTask 将任务分派到 runner
func AssignTask(dbSess *db.Session, orgGuid string, task models.Task, tpl *models.Template) {
	logger := logs.Get().WithField("action", "AssignTask").WithField("taskId", task.Guid)

	if task.Status != consts.TaskPending {
		logger.Errorf("unexpected task status '%s'", task.Status)
		return
	}

	updateTask := func(t *models.Task) error {
		if _, err := dbSess.Model(&models.Task{}).Update(t); err != nil {
			logger.Errorf("update task error: %v", err)
			return err
		}
		return nil
	}

	// 更新任务为 assigning 状态
	task.Status = consts.TaskAssigning
	if err := updateTask(&task); err != nil {
		return
	}

	now := time.Now()
	task.StartAt = &now
	resp, retry, err := doAssignTask(orgGuid, &task, tpl)
	if err == nil && resp.Error != "" {
		err = fmt.Errorf(resp.Error)
	}

	if err != nil {
		logger.Errorf("doAssignTask error: %v, retry=%v", err, retry)

		if retry {
			task.Status = consts.TaskPending // 恢复任务为 pending 状态，等待重试
			task.StatusDetail = ""
			task.StartAt = nil
			updateTask(&task)
		} else {
			// 记录任务下发失败
			task.Status = consts.TaskFailed
			task.StatusDetail = err.Error()
			updateTask(&task)
		}
	} else {
		task.Status = consts.TaskRunning
		task.StatusDetail = ""
		task.BackendInfo = getBackendInfo(task.BackendInfo, resp.Id)
		updateTask(&task)
	}

}

func doAssignTask(orgGuid string, task *models.Task, tpl *models.Template) (
	resp *runnerResp, retry bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			logs.Get().Debugf("stack: %s", debug.Stack())
			retry = task.Status == consts.TaskAssigning
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	logger := logs.Get().WithField("action", "doAssignTask").WithField("taskId", task.Guid)

	//// 组装请求
	repoAddr := tpl.RepoAddr
	if u, err := url.Parse(repoAddr); err != nil {
		return nil, false, fmt.Errorf("parse repo addr error: %v", err)
	} else if u.User == nil { // 如果 repoAddr 没有带认证信息则使用配置文件中的默认 vcs 认证信息
		defaultVcs := configs.Get().Gitlab
		u.User = url.UserPassword(defaultVcs.Username, defaultVcs.Token)
		repoAddr = u.String()
	}

	taskBackend := make(map[string]interface{}, 0)
	_ = json.Unmarshal(task.BackendInfo, &taskBackend)

	//有状态云模版，以模版ID为路径，无状态云模版，以模版ID + 作业ID 为路径
	var stateKey string
	if tpl.SaveState {
		stateKey = fmt.Sprintf("%s/%s.tfstate", orgGuid, tpl.Guid)
	} else {
		stateKey = fmt.Sprintf("%s/%s/%s.tfstate", orgGuid, tpl.Guid, task.Guid)
	}

	data := map[string]interface{}{
		"repo":          repoAddr,
		"repo_branch":   tpl.RepoBranch,
		"repo_commit":   task.CommitId,
		"template_uuid": tpl.Guid,
		"task_id":       task.Guid,
		"state_store": map[string]interface{}{
			"save_state":            tpl.SaveState,
			"backend":               "consul",
			"scheme":                "http",
			"state_key":             stateKey,
			"state_backend_address": configs.Get().Consul.Address,
		},
		"env":     runningTaskEnvParam(tpl, task.CtServiceId, task),
		"varfile": tpl.Varfile,
		"mode":    task.TaskType,
		"extra":   tpl.Extra,
	}

	header := &http.Header{}
	header.Set("Content-Type", "application/json")

	runnerAddr := taskBackend["backend_url"]
	addr := fmt.Sprintf("%s%s", runnerAddr, consts.RunnerRunTaskURL)
	logger.Infof("assign task to '%s'", addr)
	logger.Debugf("post data: %s", utils.MustJSON(data))

	// 向 runner 下发 task
	resp, err = requestRunnerRunTask(addr, header, data)
	if err != nil {
		return resp, true, fmt.Errorf("request runner failed: %v", err)
	}
	logger.Infof("runner response: %#v", resp)
	return resp, false, err
}

type runnerResp struct {
	Id    string `json:"id" form:"id" `
	Code  string `json:"code" form:"code" `
	Error string `json:"err" form:"err" `
}

func requestRunnerRunTask(url string, header *http.Header, data interface{}) (*runnerResp, error) {
	respData, err := utils.HttpService(url, "POST", header, data, 20, 5)
	if err != nil {
		return nil, err
	}

	resp := runnerResp{}
	if err := json.Unmarshal(respData, &resp); err != nil {
		return nil, fmt.Errorf("unexpected response: %s", respData)
	}
	return &resp, nil
}

// WaitTask 等待任务结束(或超时)，返回任务最新状态
func WaitTask(ctx context.Context, taskId string, tpl *models.Template, dbSess *db.Session) (status string, err error) {
	logger := logs.Get().WithField("action", "WaitTask").WithField("taskId", taskId)
	err = utils.RetryFunc(10, func(retryN int) (bool, error) {
		status, err = doPullTaskStatus(ctx, taskId, tpl, dbSess)
		if err != nil {
			logger.Errorf("pull task status error: %v, retry=%d", err, retryN)
			return true, err
		}
		return false, nil
	})
	return status, err
}

// PullTaskStatus 同步任务最新状态，直到任务结束
// 该函数允许重复调用，即使任务己结束 (runner 会在本地保存近期(约7天)任务执行信息)
func doPullTaskStatus(ctx context.Context, taskId string, tpl *models.Template, dbSess *db.Session) (
	taskStatus string, err error) {
	logger := logs.Get().WithField("action", "PullTaskState").WithField("taskId", taskId)

	// 获取 task 最新状态
	task, err := GetTaskByGuid(dbSess, taskId)
	if err != nil {
		logger.Errorf("query task err: %v", err)
		return "", err
	}
	taskStatus = task.Status

	taskBackend := make(map[string]interface{}, 0)
	_ = json.Unmarshal(task.BackendInfo, &taskBackend)
	runnerAddr := taskBackend["backend_url"]

	params := url.Values{}
	params.Add("templateUuid", task.TemplateGuid)
	params.Add("taskId", task.Guid)
	params.Add("containerId", fmt.Sprintf("%s", taskBackend["container_id"]))

	wsConn, err := utils.WebsocketDail(fmt.Sprintf("%s", runnerAddr), consts.RunnerTaskStateURL, params)
	if err != nil {
		logger.Errorln(err)
		return taskStatus, err
	}
	defer func() {
		wsConn.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
			time.Now().Add(time.Second),
		)
		wsConn.Close()
	}()

	messageChan := make(chan *runner.TaskStatusMessage)
	readErrChan := make(chan error)
	readMessage := func() {
		defer close(messageChan)

		for {
			message := runner.TaskStatusMessage{}
			if err := wsConn.ReadJSON(&message); err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
					logger.Tracef("read message error: %v", err)
				} else {
					logger.Errorf("read message error: %v", err)
					readErrChan <- err
				}
				break
			} else {
				messageChan <- &message
			}
		}
	}
	go readMessage()

	deadline := task.StartAt.Add(time.Duration(tpl.Timeout) * time.Second)
	now := time.Now()
	var timer *time.Timer
	if deadline.Before(now) {
		timer = time.NewTimer(time.Second * 10)
	} else {
		timer = time.NewTimer(deadline.Sub(now))
	}
	var lastMessage *runner.TaskStatusMessage

forLoop:
	for {
		select {
		case msg := <-messageChan:
			if msg == nil { // closed
				break forLoop
			}

			lastMessage = msg
			if lastMessage.Status == consts.DockerStatusExited {
				if msg.StatusCode == 0 {
					taskStatus = consts.TaskComplete
				} else {
					taskStatus = consts.TaskFailed
				}
				break
			}

		case err = <-readErrChan:
			return taskStatus, fmt.Errorf("read message error: %v", err)

		case <-ctx.Done():
			logger.Infof("context done with: %v", ctx.Err())
			return taskStatus, nil

		case <-timer.C:
			taskStatus = consts.TaskTimeout
			break forLoop
		}
	}

	if taskStatus != consts.TaskRunning && len(lastMessage.LogContent) > 0 {
		if err := writeTaskLog(lastMessage.LogContent,
			fmt.Sprintf("%s", taskBackend["log_file"]), 0); err != nil {
			logger.Errorf("write task log error: %v", err)
			logger.Infof("task log content: %v", lastMessage.LogContent)
		}
	}

	updateM := map[string]interface{}{
		"status": taskStatus,
		"end_at": time.Now(),
	}
	updateM["end_at"] = time.Now()
	if taskStatus != consts.TaskRunning && task.Source == consts.WorkFlow {
		k := kafka.Get()
		if err := k.ConnAndSend(k.GenerateKafkaContent(task.TransactionId, taskStatus)); err != nil {
			logger.Errorf("kafka send error: %v", err)
		}
	}

	//更新task状态
	if _, err := dbSess.Model(&models.Task{}).Where("id = ?", task.Id).Update(updateM); err != nil {
		return taskStatus, err
	}
	return taskStatus, nil
}
