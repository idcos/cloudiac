package apps

import (
	"cloudiac/consts"
	"cloudiac/consts/e"
	"cloudiac/libs/ctx"
	"cloudiac/models"
	"cloudiac/models/forms"
	"cloudiac/services"
)

type TplAccessToken struct {
	models.Template
	TplGuid     string `json:"tplGuid" form:"tplGuid" `
	AccessToken string `json:"accessToken" form:"accessToken" `
	Action      string `json:"action" form:"action" `
}

func AccessTokenHandler(c *ctx.ServiceCtx, form forms.AccessTokenHandler) (interface{}, e.Error) {
	tplInfo := TplAccessToken{}
	if err := services.AccessTokenHandlerQuery(c.DB().Debug(), form.AccessToken).First(&tplInfo); err != nil {
		return nil, e.New(e.DBError, err)
	}
	taskForm := &forms.CreateTaskForm{
		CtServiceIp:   tplInfo.DefaultRunnerAddr,
		CtServicePort: tplInfo.DefaultRunnerPort,
		CtServiceId:   tplInfo.DefaultRunnerServiceId,
		TemplateId:    tplInfo.Id,
		TemplateGuid:  tplInfo.Guid,
		TaskType:      tplInfo.Action,
	}
	c.OrgId = tplInfo.OrgId
	//判断回调动作
	switch tplInfo.Action {
	case consts.TaskPlan:
		return CreateTask(c, taskForm)
	case consts.TaskApply:
		return CreateTask(c, taskForm)
	}
	return nil, nil
}
