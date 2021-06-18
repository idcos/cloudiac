package services

import (
	"cloudiac/configs"
	"cloudiac/consts/e"
	"cloudiac/services/logstorage"
	"cloudiac/services/sshkey"
	"cloudiac/utils/kafka"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"net/http"
	"net/url"
	"time"

	"cloudiac/consts"
	"cloudiac/libs/db"
	"cloudiac/models"
	"cloudiac/runner"
	"cloudiac/utils"
	"cloudiac/utils/logs"
)

// StartTask 启动任务(任务下发后即退出)
func StartTask(dbSess *db.Session, task *models.Task) (err error) {
	logger := logs.Get().WithField("action", "StartTask").WithField("taskId", task.Guid)
	if task.Status != consts.TaskAssigning {
		return fmt.Errorf("unexpected task status '%s'", task.Status)
	}

	logger.Infof("start task %v", task.Guid)
	if err := assignTask(dbSess, task); err != nil {
		logger.Errorf("AssignTask error: %v", err)
		return err
	}

	return nil
}

// assignTask 将任务分派到 runner，并更新任务状态
func assignTask(dbSess *db.Session, task *models.Task) error {
	logger := logs.Get().WithField("action", "AssignTask").WithField("taskId", task.Guid)

	updateTask := func(t *models.Task) error {
		if _, err := dbSess.Model(&models.Task{}).Update(t); err != nil {
			err = fmt.Errorf("update task error: %v", err)
			logger.Errorln(err)
			return err
		}
		return nil
	}

	tpl := models.Template{}
	if err := dbSess.Where("id = ?", task.TemplateId).First(&tpl); err != nil {
		if e.IsRecordNotFound(err) {
			task.Status = consts.TaskFailed
			task.StatusDetail = fmt.Errorf("tplId '%d' not found", task.TemplateId).Error()
			_ = updateTask(task)
		}
		logger.Errorf("query template '%d' error: %v", task.TemplateId, err)
		return errors.Wrap(err, "query template error")
	}

	org, er := GetOrganizationById(dbSess, tpl.OrgId)
	if er != nil {
		if e.IsRecordNotFound(er) {
			task.Status = consts.TaskFailed
			task.StatusDetail = fmt.Errorf("orgId '%d' not found", tpl.OrgId).Error()
			_ = updateTask(task)
		}
		return errors.Wrap(er, "query org error")
	}

	vcs, er := QueryVcsByVcsId(tpl.VcsId, dbSess)
	if er != nil {
		if e.IsRecordNotFound(er) {
			task.Status = consts.TaskFailed
			task.StatusDetail = fmt.Errorf("vcsId '%d' not found", tpl.VcsId).Error()
			_ = updateTask(task)
		}
		return errors.Wrap(er, "query vcs error")
	}

	// 更新任务为 assigning 状态
	task.Status = consts.TaskAssigning
	if err := updateTask(task); err != nil {
		return err
	}

	now := time.Now()
	task.StartAt = &now
	logger.Debugf("assign task")
	resp, retry, err := doAssignTask(org.Guid, vcs, &tpl, task)
	if err == nil && resp.Error != "" {
		err = fmt.Errorf(resp.Error)
	}

	if err != nil {
		if retry {
			task.Status = consts.TaskPending // 恢复任务为 pending 状态，等待重试
			task.StatusDetail = ""
			task.StartAt = nil
			_ = updateTask(task)
		} else {
			// 记录任务下发失败
			task.Status = consts.TaskFailed
			task.StatusDetail = err.Error()
			_ = updateTask(task)
		}
	} else {
		task.Status = consts.TaskRunning
		task.StatusDetail = ""
		task.BackendInfo.ContainerId = resp.Id
		_ = updateTask(task)
	}
	return err
}

func doAssignTask(orgId string, vcs *models.Vcs, tpl *models.Template, task *models.Task) (
	resp *runnerResp, retry bool, err error) {
	logger := logs.Get().WithField("action", "doAssignTask").WithField("taskId", task.Guid)

	privateKey, err := sshkey.LoadPrivateKeyPem()
	if err != nil {
		return nil, true, err
	}

	//// 组装请求
	repoAddr := tpl.RepoAddr
	if u, err := url.Parse(repoAddr); err != nil {
		return nil, false, fmt.Errorf("parse repo addr error: %v", err)
	} else if u.User == nil && vcs.VcsToken != "" { // 如果 tpl.repoAddr 中没有提供认证信息则使用 vcs 的 token
		// 使用空 username 在容器中会出现认证失败的问题，暂时将 username 设置为 "token"
		u.User = url.UserPassword("token", vcs.VcsToken)
		repoAddr = u.String()
	}

	backend := task.BackendInfo

	var stateKey string
	if tpl.SaveState {
		// 有状态云模版，以模版ID为路径，
		stateKey = fmt.Sprintf("%s/%s.tfstate", orgId, tpl.Guid)
	} else {
		// 无状态云模版，以模版ID + 作业ID 为路径
		stateKey = fmt.Sprintf("%s/%s/%s.tfstate", orgId, tpl.Guid, task.Guid)
	}

	data := map[string]interface{}{
		"repo":          repoAddr,
		"repo_branch":   tpl.RepoBranch,
		"repo_commit":   task.CommitId,
		"template_uuid": tpl.Guid,
		"task_id":       task.Guid,
		"state_store": runner.StateStore{
			SaveState:           tpl.SaveState,
			Backend:             "consul",
			Scheme:              "http",
			StateKey:            stateKey,
			StateBackendAddress: configs.Get().Consul.Address,
			Lock:                true,
		},
		"env":      runningTaskEnvParam(tpl, task.CtServiceId, task),
		"varfile":  tpl.Varfile,
		"mode":     task.TaskType,
		"extra":    tpl.Extra,
		"playbook": tpl.Playbook,

		"privateKey": string(privateKey),
	}

	header := &http.Header{}
	header.Set("Content-Type", "application/json")

	addr := fmt.Sprintf("%s%s", backend.BackendUrl, consts.RunnerRunTaskURL)
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
	respData, err := utils.HttpService(url, "POST", header, data, 1, 5)
	if err != nil {
		return nil, err
	}

	resp := runnerResp{}
	if err := json.Unmarshal(respData, &resp); err != nil {
		return nil, fmt.Errorf("unexpected response: %s", respData)
	}
	return &resp, nil
}

// WaitTaskResult 等待任务结束(包括超时)，返回任务最新状态
// 该函数会更新任务状态、日志等到 db
// param: taskDeadline 任务超时时间，达到这个时间后任务会被置为 timeout 状态
func WaitTaskResult(ctx context.Context, dbSess *db.Session, task *models.Task, taskDeadline time.Time) (status string, err error) {
	logger := logs.Get().WithField("action", "WaitTaskResult").WithField("taskId", task.Guid)

	// 当前版本实现中需要 portal 主动连接到 runner 获取状态
	err = utils.RetryFunc(0, time.Second*10, func(retryN int) (bool, error) {
		status, err = doPullTaskStatus(ctx, dbSess, task.Guid, taskDeadline)
		if err != nil {
			logger.Errorf("pull task status error: %v, retry=%d", err, retryN)
			return true, err
		}
		return false, nil
	})
	if err != nil {
		return "", err
	}

	updateM := map[string]interface{}{
		"status": status,
		"end_at": time.Now(),
	}
	updateM["end_at"] = time.Now()
	if status != consts.TaskRunning && task.Source == consts.WorkFlow {
		k := kafka.Get()
		if err := k.ConnAndSend(k.GenerateKafkaContent(task.TransactionId, status)); err != nil {
			logger.Errorf("kafka send error: %v", err)
		}
	}

	//更新 task 状态
	if _, err := dbSess.Model(&models.Task{}).
		Where("id = ?", task.Id).Update(updateM); err != nil {
		return status, err
	}

	if status == consts.TaskComplete {
		// 解析日志输出，更新资源变更信息
		tfInfo := ParseTfOutput(task.BackendInfo.LogFile)
		models.UpdateAttr(dbSess.Where("id = ?", task.Id), &models.Task{}, tfInfo)
	}

	return status, err
}

// PullTaskStatus 同步任务最新状态，直到任务结束(或 ctx cancel)
// 该函数允许重复调用，即使任务己结束 (runner 会在本地保存近期(约7天)任务执行信息)，如果任务结束则写入全量日志到存储
func doPullTaskStatus(ctx context.Context, dbSess *db.Session, taskId string, deadline time.Time) (
	taskStatus string, err error) {
	logger := logs.Get().WithField("action", "PullTaskState").WithField("taskId", taskId)

	// 获取 task 最新状态
	task, err := GetTaskByGuid(dbSess, taskId)
	if err != nil {
		logger.Errorf("query task err: %v", err)
		return "", err
	}
	taskStatus = task.Status

	backend := task.BackendInfo
	runnerAddr := backend.BackendUrl
	params := url.Values{}
	params.Add("templateId", task.TemplateGuid)
	params.Add("taskId", task.Guid)
	params.Add("containerId", fmt.Sprintf("%s", backend.ContainerId))
	wsConn, err := utils.WebsocketDail(fmt.Sprintf("%s", runnerAddr), consts.RunnerTaskStateURL, params)
	if err != nil {
		logger.Errorf("connect error: %v", err)
		return taskStatus, err
	}
	defer utils.WebsocketClose(wsConn)

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

	now := time.Now()
	var timer *time.Timer
	if deadline.Before(now) {
		// 即使任务己超时也保证进行一次状态获取
		timer = time.NewTimer(time.Second)
	} else {
		timer = time.NewTimer(deadline.Sub(now))
	}
	var lastMessage *runner.TaskStatusMessage

	logger.Debugf("pulling task status ...")
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

	logger.Debugf("pull task status done, status=%s", taskStatus)

	if taskStatus != consts.TaskRunning && len(lastMessage.LogContent) > 0 {
		path := task.BackendInfo.LogFile
		if err := logstorage.Get().Write(path, lastMessage.LogContent); err != nil {
			logger.WithField("path", path).Errorf("write task log error: %v", err)
			logger.Infof("task log content: %s", lastMessage.LogContent)
		}
	}

	return taskStatus, nil
}

func TaskDeadline(dbSess *db.Session, taskId string) (deadline time.Time, err error) {
	result := struct {
		StartAt time.Time
		Timeout int64
	}{}
	err = dbSess.Raw("SELECT tpl.timeout, task.start_at FROM iac_template AS tpl "+
		"JOIN iac_task AS task ON task.template_guid = tpl.guid "+
		"WHERE task.guid = ?", taskId).Scan(&result)
	if err != nil {
		return
	}
	return result.StartAt.Add(time.Duration(result.Timeout) * time.Second), nil
}
