package apps

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
)

type TplAccessToken struct {
	models.Template
	TplGuid     string `json:"tplGuid" form:"tplGuid" `
	AccessToken string `json:"accessToken" form:"accessToken" `
	Action      string `json:"action" form:"action" `
}

func AccessTokenHandler(c *ctx.ServiceCtx, form forms.AccessTokenHandler) (interface{}, e.Error) {
	// TODO 待实现
	//tplInfo := TplAccessToken{}
	//if err := servic待AccessTokenHandlerQuery(c.DB().Debug(), form.AccessToken).First(&tplInfo); err != nil {
	//	return nil, e.New(e.DBError, err)
	//}
	//taskForm := &forms.CreateTaskForm{
	//	CtServiceIp:   tplInfo.DefaultRunnerAddr,
	//	CtServicePort: tplInfo.DefaultRunnerPort,
	//	CtServiceId:   tplInfo.DefaultRunnerServiceId,
	//	TemplateId:    tplInfo.Id,
	//	TemplateGuid:  tplInfo.Guid,
	//	TaskType:      tplInfo.Action,
	//}
	//c.OrgId = tplInfo.OrgId
	////判断回调动作
	//switch tplInfo.Action {
	//case consts.TaskPlan:
	//	return CreateTask(c, taskForm)
	//case consts.TaskApply:
	//	return CreateTask(c, taskForm)
	//}
	return nil, nil
}
