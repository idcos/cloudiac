package services

import (
	"bufio"
	"bytes"
	"cloudiac/consts"
	"cloudiac/consts/e"
	"cloudiac/libs/db"
	"cloudiac/models"
	"cloudiac/models/forms"
	"cloudiac/runner"
	"cloudiac/services/logstorage"
	"cloudiac/utils"
	"cloudiac/utils/logs"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
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

func GetTaskByGuid(tx *db.Session, guid string) (*models.Task, e.Error) {
	o := models.Task{}
	if err := tx.Where("guid = ?", guid).First(&o); err != nil {
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

func GetLastTaskByTemplateGuid(tx *db.Session, tplGuid string) (*models.Task, e.Error) {
	task := &models.Task{}
	if err := tx.Table(models.Task{}.TableName()).
		Where("template_guid = ?", tplGuid).
		Where("task_type in (?)", []string{
			consts.TaskApply,
			consts.TaskDestroy,
		}).
		Last(task); err != nil {
		return nil, e.New(e.DBError, err)
	}
	return task, nil
}

func TaskStateList(query *db.Session, tplGuid string) (interface{}, e.Error) {
	stateList := make([]string, 0)
	var reader io.Reader
	lastTask, err := GetLastTaskByTemplateGuid(query, tplGuid)
	if err != nil {
		if e.IsRecordNotFound(err) {
			return stateList, nil
		}
		return nil, err
	}
	taskPath := runner.GetTaskWorkDir(lastTask.TemplateGuid, lastTask.Guid)
	path := filepath.Join(taskPath, consts.TerraformStateListName)
	if content, err := logstorage.Get().Read(path); err != nil {
		if e.IsRecordNotFound(err) {
			return stateList, nil
		}
		return nil, e.New(e.TaskNotExists, err)
	} else {
		reader = bytes.NewBuffer(content)
	}
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		stateList = append(stateList, scanner.Text())
	}
	return stateList, nil
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

func runningTaskEnvParam(tpl *models.Template, runnerId string, task *models.Task) interface{} {
	tplVars := make([]forms.Var, 0)
	taskVars := make([]forms.VarOpen, 0)
	param := make(map[string]interface{})

	tplVarsByte, _ := tpl.Vars.MarshalJSON()
	taskVarsByte, _ := task.SourceVars.MarshalJSON()

	if !tpl.Vars.IsNull() {
		_ = json.Unmarshal(tplVarsByte, &tplVars)
	}

	if !task.SourceVars.IsNull() {
		_ = json.Unmarshal(taskVarsByte, &taskVars)
	}

	tplVars = append(tplVars, resourceEnvParam(runnerId, tpl.OrgId)...)
	for _, v := range varsDuplicateRemoval(taskVars, tplVars) {
		if v.Key == "" {
			continue
		}
		if v.Type == consts.Terraform && !strings.HasPrefix(v.Key, consts.TerraformVar) {
			v.Key = fmt.Sprintf("%s%s", consts.TerraformVar, v.Key)
		}
		if v.IsSecret != nil && *v.IsSecret {
			param[v.Key] = utils.AesDecrypt(v.Value)
		} else {
			param[v.Key] = v.Value
		}
	}
	return param
}

func varsDuplicateRemoval(taskVars []forms.VarOpen, tplVars []forms.Var) []forms.Var {
	if taskVars == nil || len(taskVars) == 0 {
		return tplVars
	}
	vars := make([]forms.Var, 0)
	//taskV := make(map[string]forms.VarOpen, 0)
	tplV := make(map[string]forms.Var, 0)
	for _, tplv := range tplVars {
		tplV[tplv.Key] = tplv
	}
	isSecret := false
	for _, taskv := range taskVars {
		if taskv.Name == "" {
			continue
		}
		if taskv.Value == "" {
			vars = append(vars, tplV[taskv.Name])
		} else {
			vars = append(vars, forms.Var{
				Key:      taskv.Name,
				Value:    taskv.Value,
				IsSecret: &isSecret,
			})
		}
	}
	return vars
}

func resourceEnvParam(runnerId string, orgId uint) []forms.Var {
	vars := make([]forms.Var, 0)
	ra := []models.ResourceAccount{}
	//org,_:=getorg
	if err := db.Get().Debug().Joins(fmt.Sprintf("left join %s as crm on %s.id = crm.resource_account_id",
		models.CtResourceMap{}.TableName(), models.ResourceAccount{}.TableName())).
		Where("crm.ct_service_id = ?", runnerId).
		Where(fmt.Sprintf("%s.status = ?", models.ResourceAccount{}.TableName()), consts.ResourceAccountEnable).
		Where(fmt.Sprintf("%s.org_id = ?", models.ResourceAccount{}.TableName()), orgId).
		Find(&ra); err != nil {
		logs.Get().Errorf("ResourceAccount db err %v: ", err)
		return nil
	}

	for _, raInfo := range ra {
		varsByte, _ := raInfo.Params.MarshalJSON()
		if !raInfo.Params.IsNull() {
			v := make([]forms.Var, 0)
			_ = json.Unmarshal(varsByte, &v)
			vars = append(vars, v...)
		}
	}

	return vars
}

func getBackendInfo(backendInfo models.JSON, containerId string) []byte {
	attr := models.Attrs{}
	_ = json.Unmarshal(backendInfo, &attr)
	attr["container_id"] = containerId
	b, _ := json.Marshal(attr)
	return b
}

var (
	planChangesLineRegex    = regexp.MustCompile(`([\d]+) to add, ([\d]+) to change, ([\d]+) to destroy`)
	applyChangesLineRegex   = regexp.MustCompile(`Apply complete! Resources: ([\d]+) added, ([\d]+) changed, ([\d]+) destroyed.`)
	destroyChangesLineRegex = regexp.MustCompile(`Destroy complete! Resources: ([\d]+) destroyed.`)
)

func ParseTfOutput(path string) map[string]interface{} {
	loggers := logs.Get()
	content, err := logstorage.Get().Read(path)
	if err != nil {
		loggers.Errorf("read log file %s: %v", path, err)
		return nil
	}

	result := make(map[string]interface{})
	rd := bufio.NewReader(bytes.NewBuffer(content))
	for {
		str, _, err := rd.ReadLine()
		if err != nil {
			if err.Error() == "EOF" {
				break
			} else {
				loggers.Error("Read Error:", err.Error())
				break
			}
		}
		LogStr := string(str)
		if strings.Contains(LogStr, "No changes. Infrastructure is up-to-date.") {
			result["add"] = "0"
			result["change"] = "0"
			result["destroy"] = "0"
			result["allowApply"] = false
			break
		} else if strings.Contains(LogStr, `Plan:`) {
			params := planChangesLineRegex.FindStringSubmatch(LogStr)
			if len(params) == 4 {
				result["add"] = params[1]
				result["change"] = params[2]
				result["destroy"] = params[3]
				result["allowApply"] = true
			}
			break
		} else if strings.Contains(LogStr, `Apply complete!`) {
			params := applyChangesLineRegex.FindStringSubmatch(LogStr)
			if len(params) == 4 {
				result["add"] = params[1]
				result["change"] = params[2]
				result["destroy"] = params[3]
				result["allowApply"] = false
			}
			break
		} else if strings.Contains(LogStr, `Destroy complete!`) {
			params := destroyChangesLineRegex.FindStringSubmatch(LogStr)
			if len(params) == 2 {
				result["add"] = "0"
				result["change"] = "0"
				result["destroy"] = params[1]
				result["allowApply"] = false
			}
			break
		}
	}
	return result
}

type sendMailQuery struct {
	models.NotificationCfg
	models.User
}

func SendMail(query *db.Session, orgId uint, task *models.Task) {
	tos := make([]string, 0)
	logger := logs.Get().WithField("action", "sendMail")
	notifier := make([]sendMailQuery, 0)
	if err := query.Debug().Table(models.NotificationCfg{}.TableName()).Where("org_id = ?", orgId).
		Joins(fmt.Sprintf("left join %s as `user` on `user`.id = %s.user_id", models.User{}.TableName(), models.NotificationCfg{}.TableName())).
		LazySelectAppend("`user`.*").
		LazySelectAppend("`iac_org_notification_cfg`.*").
		Scan(&notifier); err != nil {
		logger.Errorf("query notifier err: %v", err)
		return
	}

	tpl, _ := GetTemplateById(query, task.TemplateId)
	for _, v := range notifier {
		user, _ := GetUserById(query, v.UserId)
		switch task.Status {
		case consts.TaskPending:
			if v.EventType == "all" {
				tos = append(tos, user.Email)
			}
		case consts.TaskComplete:
			if v.EventType == "all" {
				tos = append(tos, user.Email)
			}
		case consts.TaskFailed:
			if v.EventType == "all" || v.EventType == "failure" {
				tos = append(tos, user.Email)
			}
		case consts.TaskTimeout:
			if v.EventType == "all" || v.EventType == "failure" {
				tos = append(tos, user.Email)
			}
		}
	}

	tos = utils.RemoveDuplicateElement(tos)
	if len(tos) == 0 {
		return
	}
	sendMail := GetMail(tos, *task, tpl)
	sendMail.SendMail()

}

func DefaultRunner(dbSess *db.Session, runnerAddr string, runnerPort uint, tplId, orgId uint) (string, uint, e.Error) {
	if runnerAddr != "" && runnerPort != 0 {
		return runnerAddr, runnerPort, nil
	}

	if tplId != 0 {
		tpl, err := GetTemplateById(dbSess, tplId)
		if err != nil {
			return "", 0, err
		}
		if tpl.DefaultRunnerAddr == "" || tpl.DefaultRunnerPort == 0 {
			return DefaultRunner(dbSess, "", 0, 0, orgId)
		}
		return tpl.DefaultRunnerAddr, tpl.DefaultRunnerPort, nil
	}

	if orgId != 0 {
		org, err := GetOrganizationById(dbSess, tplId)
		if err != nil {
			return "", 0, err
		}
		return org.DefaultRunnerAddr, org.DefaultRunnerPort, nil
	}

	return "", 0, nil
}
