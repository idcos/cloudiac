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
	"reflect"
)

func TaskLogSSEGetPath(c *ctx.ServiceCtx, taskGuid string) string {
	task, err := services.GetTaskByGuid(c.DB(), taskGuid)
	if err != nil {
		logs.Get().Error(err)
	}
	taskBackend := make(map[string]interface{}, 0)
	_ = json.Unmarshal(task.BackendInfo, &taskBackend)
	return taskBackend["log_file"].(string)
}

func CreateTaskOpen(c *ctx.ServiceCtx, form forms.CreateTaskOpenForm) (interface{}, e.Error) {
	tx := c.DB().Debug()
	guid := utils.GenGuid("run")
	conf := configs.Get()
	//todo 如果cmp寻址提供runner信息需要使用寻址的runner进行作业执行
	logPath := fmt.Sprintf("%s/%s/%s", conf.Task.LogPath, form.TemplateGuid, guid)

	//根据模板GUID获取模板id
	tpl, err := services.GetTemplateByGuid(tx, form.TemplateGuid)
	if err != nil {
		return nil, err
	}
	b, _ := json.Marshal(map[string]interface{}{
		"backend_url": fmt.Sprintf("http://%s:%d/api/v1", tpl.DefaultRunnerAddr, tpl.DefaultRunnerPort),
		"ctServiceId": conf.Consul.ServiceID,
		"log_file":    logPath,
		"log_offset":  0,
	})

	vars := GetResourceAccount(form.Account, form.Vars, tpl.TplType)
	jsons, _ := json.Marshal(vars)

	task, err := services.CreateTask(tx, models.Task{
		TemplateGuid:  form.TemplateGuid,
		TaskType:      consts.TaskApply,
		Guid:          guid,
		CommitId:      form.CommitId,
		Source:        form.Source,
		SourceVars:    models.JSON(jsons),
		CtServiceId:   conf.Consul.ServiceID,
		BackendInfo:   models.JSON(b),
		TemplateId:    tpl.Id,
		TransactionId: form.TransactionId,
		Creator:       c.UserId,
	})
	if err != nil {
		return nil, err
	}
	//go services.RunTaskToRunning(task, c.DB().Debug(), org.Guid)
	//go services.StartTask(c.DB(), *task)

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
