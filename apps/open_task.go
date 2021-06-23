package apps

import (
	"cloudiac/configs"
	"cloudiac/consts"
	"cloudiac/consts/e"
	"cloudiac/libs/ctx"
	"cloudiac/models"
	"cloudiac/models/forms"
	"cloudiac/services"
	"cloudiac/utils"
	"cloudiac/utils/logs"
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
)

func TaskLogSSEGetPath(c *ctx.ServiceCtx, taskGuid string) string {
	task, err := services.GetTaskByGuid(c.DB(), taskGuid)
	if err != nil {
		logs.Get().Error(err)
	}
	return task.BackendInfo.LogFile
}

func CreateTaskOpen(c *ctx.ServiceCtx, form forms.CreateTaskOpenForm) (interface{}, e.Error) {
	dbSess := c.DB().Debug()
	guid := utils.GenGuid("run")
	conf := configs.Get()
	logPath := filepath.Join(form.TemplateGuid, guid, consts.TaskLogName)

	//根据模板GUID获取模板id
	tpl, err := services.GetTemplateByGuid(dbSess, form.TemplateGuid)
	if err != nil {
		return nil, err
	}

	runnerAddr, runnerPort, err := services.DefaultRunner(dbSess, "", 0, tpl.Id, tpl.OrgId)
	if err != nil {
		return nil, err
	}

	backend := models.TaskBackendInfo{
		BackendUrl:  fmt.Sprintf("http://%s:%d/api/v1", runnerAddr, runnerPort),
		CtServiceId: conf.Consul.ServiceID,
		LogFile:     logPath,
	}

	vars := GetResourceAccount(form.Account, form.Vars, tpl.TplType)
	jsons, _ := json.Marshal(vars)
	task, err := services.CreateTask(dbSess, models.Task{
		TemplateGuid:  form.TemplateGuid,
		TaskType:      consts.TaskApply,
		Guid:          guid,
		CommitId:      form.CommitId,
		Source:        form.Source,
		SourceVars:    models.JSON(jsons),
		CtServiceId:   conf.Consul.ServiceID,
		BackendInfo:   &backend,
		TemplateId:    tpl.Id,
		TransactionId: form.TransactionId,
		Creator:       c.UserId,
		Status:        consts.TaskPending,
	})

	if err != nil {
		return nil, err
	}

	return task, nil
}

func GetResourceAccount(account forms.Account, open []forms.VarOpen, accountType string) (vars []forms.VarOpen) {
	am := consts.AccountMap[accountType]
	types := reflect.TypeOf(&account).Elem()   //通过反射获取type定义
	values := reflect.ValueOf(&account).Elem() //通过反射获取type定义
	for i := 0; i < types.NumField(); i++ {
		if _, ok := am[types.Field(i).Tag.Get("json")]; !ok {
			continue
		}
		vars = append(vars, forms.VarOpen{
			Name:  am[types.Field(i).Tag.Get("json")],
			Value: values.Field(i).Interface().(string),
		})
	}
	vars = append(vars, open...)
	return

}
