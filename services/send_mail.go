package services

import (
	"cloudiac/consts"
	"cloudiac/models"
	"cloudiac/utils/mail"
	"fmt"
)

type sendMail struct {
	Tos  []string        `json:"tos" form:"tos" `
	Task models.Task     `json:"task" form:"task" `
	Tpl  models.Template `json:"tpl" form:"tpl" `
}

func (sm *sendMail) SendMail() {
	subject := fmt.Sprintf("作业运行%s", consts.StatusTranslation[sm.Task.Status])
	//"您已在云模板:%s下成功名称为:%s的plan作业"
	//"【%s】<br>【%s】[%s][P%d]<tr><td>Metric: </td><td>%s</td></tr><tr><td>Tags: </td><td>%s</td></tr><tr><td>Strategy: </td><td>%s</td></tr><tr><td>Note: </td><td>%s</td></tr><tr><td>Current: </td><td>%d/%d</td></tr><tr><td>Time: </td><td>%s</td></tr></table><br><br>",
	content := fmt.Sprintf(
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
		sm.Tpl.Guid,
		sm.Task.Guid,
		sm.Task.TaskType,
		sm.Task.Status,
		sm.Task.CtServiceId,
		sm.Task.CommitId,
		sm.Tpl.RepoBranch,
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
