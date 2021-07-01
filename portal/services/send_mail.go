package services

import (
	"bytes"
	"cloudiac/portal/consts"
	"cloudiac/portal/models"
	"cloudiac/utils/logs"
	"cloudiac/utils/mail"
	"fmt"
	"html/template"
)

type sendMail struct {
	Tos  []string        `json:"tos" form:"tos" `
	Task models.Task     `json:"task" form:"task" `
	Tpl  models.Template `json:"tpl" form:"tpl" `
}

func (sm *sendMail) SendMail() {
	logger := logs.Get()
	tmpl, err := template.New("sendMail").Parse("<table>" +
		"<tr><td>模板名称: </td><td>{{.Name}}</td></tr>" +
		"<tr><td>模板 Id: </td><td>{{.Id}}</td></tr>" +
		"<tr><td>作业 Id: </td><td>{{.TaskId}}</td></tr>" +
		"<tr><td>作业类型: </td><td>{{.TaskType}}</td></tr>" +
		"<tr><td>作业状态: </td><td>{{.Status}}</td></tr>" +
		"<tr><td>runnerId: </td><td>{{.RunnerId}}</td></tr>" +
		"<tr><td>CommitId: </td><td>{{.CommitId}}</td></tr>" +
		"<tr><td>Revision: </td><td>{{.RepoRevision}}</td></tr>" +
		"</table>")
	if err != nil {
		logger.Errorf("get template %+v", err)
		return
	}
	buffer := bytes.NewBuffer(nil)
	if err := tmpl.Execute(buffer, struct {
		RunnerId string
		CommitId string
		TaskType string
		Status   string
		TaskId   models.Id
		TaskName string
		models.Template
	}{
		RunnerId: sm.Task.RunnerId,
		CommitId: sm.Task.CommitId,
		//TaskType:    sm.Task.TaskType,
		Status:   sm.Task.Status,
		TaskId:   sm.Task.Id,
		TaskName: sm.Task.Name,
		Template: sm.Tpl,
	}); err != nil {
		logger.Errorf("get template %+v", err)
		return
	}

	subject := fmt.Sprintf("作业运行%s", consts.StatusTranslation[sm.Task.Status])
	//"您已在云模板:%s下成功名称为:%s的plan作业"
	//"【%s】<br>【%s】[%s][P%d]<tr><td>Metric: </td><td>%s</td></tr><tr><td>Tags: </td><td>%s</td></tr><tr><td>Strategy: </td><td>%s</td></tr><tr><td>Note: </td><td>%s</td></tr><tr><td>Current: </td><td>%d/%d</td></tr><tr><td>Time: </td><td>%s</td></tr></table><br><br>",
	content := string(buffer.Bytes())
	fmt.Printf(
		"<table>"+
			"<tr><td>模板名称: </td><td>%s</td></tr>"+
			"<tr><td>模板guid: </td><td>%s</td></tr>"+
			"<tr><td>作业guid: </td><td>%s</td></tr>"+
			"<tr><td>作业类型: </td><td>%s</td></tr>"+
			"<tr><td>作业状态: </td><td>%s</td></tr>"+
			"<tr><td>CtRunnerId: </td><td>%s</td></tr>"+
			"<tr><td>CommitId: </td><td>%s</td></tr>"+
			"<tr><td>Branch: </td><td>%s</td></tr>"+
			"</table>",
		sm.Tpl.Name,
		sm.Tpl.Id,
		sm.Task.Id,
		//sm.Task.TaskType,
		sm.Task.Status,
		sm.Task.RunnerId,
		sm.Task.CommitId,
		sm.Tpl.RepoRevision,
	)
	_ = mail.SendMail(sm.Tos, subject, content)
}

func GetMail(tos []string, task models.Task, template models.Template) sendMail {
	return sendMail{
		Tos:  tos,
		Task: task,
		Tpl:  template,
	}
}
