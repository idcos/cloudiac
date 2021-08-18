// Copyright 2021 CloudJ Company Limited. All rights reserved.

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

type SendMail struct {
	Tos  []string        `json:"tos" form:"tos" `
	Task models.Task     `json:"task" form:"task" `
	Tpl  models.Template `json:"tpl" form:"tpl" `
}

func (sm *SendMail) SendMail() {
	logger := logs.Get()
	tmpl, err := template.New("SendMail").Parse("<table>" +
		"<tr><td>模板名称: </td><td>{{.Name}}</td></tr>" +
		"<tr><td>模板 Id: </td><td>{{.Id}}</td></tr>" +
		"<tr><td>作业 Id: </td><td>{{.Id}}</td></tr>" +
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
		sm.Task.Type,
		sm.Task.Status,
		sm.Task.RunnerId,
		sm.Task.CommitId,
		sm.Tpl.RepoRevision,
	)
	_ = mail.SendMail(sm.Tos, subject, content)
}

func GetMail(tos []string, task models.Task, template models.Template) SendMail {
	return SendMail{
		Tos:  tos,
		Task: task,
		Tpl:  template,
	}
}

//func SendTaskMail(dbSess *db.Session, taskId models.Id) {
//	tos := make([]string, 0)
//	logger := logs.Get().WithField("action", "sendMail")
//	notifier := make([]sendMailQuery, 0)
//	if err := query.Debug().Table(models.Notification{}.TableName()).Where("org_id = ?", orgId).
//		Joins(fmt.Sprintf("left join %s as `user` on `user`.id = %s.user_id", models.User{}.TableName(), models.Notification{}.TableName())).
//		LazySelectAppend("`user`.*").
//		LazySelectAppend("`iac_org_notification_cfg`.*").
//		Scan(&notifier); err != nil {
//		logger.Errorf("query notifier err: %v", err)
//		return
//	}
//
//	tpl, _ := GetTemplateById(query, task.TemplateId)
//	for _, v := range notifier {
//		user, _ := GetUserById(query, v.UserId)
//		switch task.Status {
//		case consts.TaskPending:
//			if v.EventType == "all" {
//				tos = append(tos, user.Email)
//			}
//		case consts.TaskComplete:
//			if v.EventType == "all" {
//				tos = append(tos, user.Email)
//			}
//		case consts.TaskFailed:
//			if v.EventType == "all" || v.EventType == "failure" {
//				tos = append(tos, user.Email)
//			}
//		case consts.TaskTimeout:
//			if v.EventType == "all" || v.EventType == "failure" {
//				tos = append(tos, user.Email)
//			}
//		}
//	}
//
//	tos = utils.RemoveDuplicateElement(tos)
//	if len(tos) == 0 {
//		return
//	}
//	sendMail := GetMail(tos, *task, tpl)
//	sendMail.SendMail()
//
//}
