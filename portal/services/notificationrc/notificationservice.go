package notificationrc

import (
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/utils"
	"cloudiac/utils/logs"
	"cloudiac/utils/mail"
	"fmt"
)

type NotificationService struct {
	OrgId          models.Id    `json:"orgId" form:"orgId" `
	ProjectId      models.Id    `json:"projectId" form:"projectId" `
	Env            *models.Env  `json:"env" form:"env" `
	Task           *models.Task `json:"task" form:"task" `
	EventFailed    bool         `json:"eventFailed" form:"eventFailed" `
	EventComplete  bool         `json:"eventComplete" form:"eventComplete" `
	EventApproving bool         `json:"eventApproving" form:"eventApproving" `
	EventRunning   bool         `json:"eventRunning" form:"eventRunning" `
}

type NotificationOptions struct {
	OrgId          models.Id    `json:"orgId" form:"orgId" `
	ProjectId      models.Id    `json:"projectId" form:"projectId" `
	Env            *models.Env  `json:"env" form:"env" `
	Task           *models.Task `json:"task" form:"task" `
	EventFailed    bool         `json:"eventFailed" form:"eventFailed" `
	EventComplete  bool         `json:"eventComplete" form:"eventComplete" `
	EventApproving bool         `json:"eventApproving" form:"eventApproving" `
	EventRunning   bool         `json:"eventRunning" form:"eventRunning" `
}

func NewNotificationService(options *NotificationOptions) NotificationService {
	return NotificationService{
		OrgId:          options.OrgId,
		ProjectId:      options.ProjectId,
		Env:            options.Env,
		Task:           options.Task,
		EventFailed:    options.EventFailed,
		EventComplete:  options.EventComplete,
		EventApproving: options.EventApproving,
		EventRunning:   options.EventRunning,
	}
}

func (ns *NotificationService) SendMessage() {
	orgNotification := make([]models.Notification, 0)
	projectNotification := make([]models.Notification, 0)
	notifications := make([]models.Notification, 0)
	dbSess := db.Get().Debug().Where("org_id = ?", ns.OrgId)
	if ns.EventFailed {
		dbSess = dbSess.Where("event_failed = ?", ns.EventFailed)
	}
	if ns.EventComplete {
		dbSess = dbSess.Where("event_complete = ?", ns.EventComplete)
	}
	if ns.EventApproving {
		dbSess = dbSess.Where("event_approving = ?", ns.EventApproving)
	}
	if ns.EventRunning {
		dbSess = dbSess.Where("event_running = ?", ns.EventRunning)
	}
	// 查询需要组织下需要通知的人
	if err := dbSess.
		Where("project_id = '' or project_id = null").
		Find(&orgNotification); err != nil {

	}
	// 查询需要项目下需要通知的人
	if err := dbSess.
		Where("project_id = ?", ns.ProjectId).
		Find(&projectNotification); err != nil {
	}

	// 将需要通知的数据进行整理
	notifications = append(notifications, orgNotification...)
	notifications = append(notifications, projectNotification...)

	for _, notification := range notifications {
		//todo 消息通知模板
		fmt.Println(notification.NotificationType, "notification.NotificationType")
		switch notification.NotificationType {
		case models.NotificationTypeDingTalk:
			fmt.Println(models.NotificationTypeDingTalk, "NotificationTypeDingTalk")
			ns.SendDingTalkMessage(notification, "```test```")
		case models.NotificationTypeWebhook:
			fmt.Println(models.NotificationTypeWebhook, "NotificationTypeWebhook")
			ns.SendWebhookMessage(notification, "```test```")
		case models.NotificationTypeWeChat:
			fmt.Println(models.NotificationTypeWeChat, "NotificationTypeWeChat")
			ns.SendWechatMessage(notification, "```NotificationTypeWeChat```\n### NotificationTypeWeChat")
		case models.NotificationTypeSlack:
			fmt.Println(models.NotificationTypeSlack, "NotificationTypeSlack")
			ns.SendSlackMessage(notification, "```test```")
		case models.NotificationTypeEmail:
			fmt.Println(models.NotificationTypeEmail, "NotificationTypeEmail")
			ns.SendEmailMessage(notification, "```test```")
		}
	}

}

func (ns *NotificationService) SendDingTalkMessage(n models.Notification, message string) {
	dingTalk := NewDingTalkRobot(n.Url, n.Secret)
	if err := dingTalk.SendMarkdownMessage("IaC Notification", message, nil, false); err != nil {
		logs.Get().Errorf("send dingtalk message err: %v", err)
	}
}

func (ns *NotificationService) SendWechatMessage(n models.Notification, message string) {
	wechat := WeChatRobot{Url: n.Url}
	if _, err := wechat.SendMarkdown(message); err != nil {
		logs.Get().Errorf("send wechat message err: %v", err)

	}
}

func (ns *NotificationService) SendWebhookMessage(n models.Notification, message string) {
	w := Webhook{Url: n.Url}
	if err := w.Send(message); err != nil {
		logs.Get().Errorf("send webhook message err: %v", err)
	}
}

func (ns *NotificationService) SendSlackMessage(n models.Notification, message string) {
	if errs := SendSlack(n.Url, Payload{Text: message, Markdown: true}); len(errs) != 0 {
		logs.Get().Errorf("send slack message err: %v", errs)
	}

}

func (ns *NotificationService) SendEmailMessage(n models.Notification, message string) {
	// 获取用户邮箱列表
	users := make([]models.User, 0)
	_ = db.Get().Where("id in (?)", n.UserIds).Find(users)

	emails := make([]string, 0)
	for _, v := range users {
		emails = append(emails, v.Email)
	}
	emails = utils.RemoveDuplicateElement(emails)
	if err := mail.SendMail(emails, "", message); err != nil {

	}
}

func GetEventToStatus(status string) (eventFailed, eventComplete, eventApproving, eventRunning bool) {
	switch status {
	case models.TaskFailed:
		return true, false, false, false
	case models.TaskComplete:
		return false, true, false, false
	case models.TaskRunning:
		return false, false, false, true
	case models.TaskApproving:
		return false, false, true, false
	case models.TaskRejected:
		return false, false, true, false
	default:
		return false, false, false, false
	}

}
