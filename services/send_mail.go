package services

import (
	"cloudiac/utils/mail"
	"fmt"
)

type sendMail struct {
	Tos      []string `json:"tos" form:"tos" `
	TplName  string   `json:"tplName" form:"tplName" `
	TaskGuid string   `json:"taskGuid" form:"taskGuid" `
	TaskType string   `json:"taskType" form:"taskType" `
}

func (sm *sendMail) SendMail() {

}

func (sm *sendMail) SendMailToSuccessTask() {
	subject := "作业运行成功"
	//"您已在云模板:%s下成功名称为:%s的plan作业"
	content := fmt.Sprintf("作业运行成功,作业id:%s", sm.TaskGuid)
	_ = mail.SendMail(sm.Tos, subject, content)

}

func (sm *sendMail) SendMailToFailTask() {
	subject := "作业运行失败"
	content := fmt.Sprintf("作业运行失败，作业id为:%s", sm.TaskGuid)
	_ = mail.SendMail(sm.Tos, subject, content)

}

func (sm *sendMail) SendMailToCreateTask() {
	subject := "作业创建成功"
	//"您已在云模板:%s下成功名称为:%s的plan作业"
	content := fmt.Sprintf("您已在云模板:%s下成功Id为:%s的%s作业", sm.TplName, sm.TaskGuid, sm.TaskType)
	_ = mail.SendMail(sm.Tos, subject, content)
}

func GetMail(tos []string, tplName, taskGuid, taskType string) sendMail {
	return sendMail{
		Tos:      tos,
		TplName:  tplName,
		TaskGuid: taskGuid,
		TaskType: taskType,
	}
}
