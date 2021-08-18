package notificationrc

import (
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/portal/services"
	"cloudiac/utils"
	"cloudiac/utils/mail"
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
	dbSess := db.Get().Where("org_id = ?", ns.OrgId).
		Where("event_failed = ?", ns.EventFailed).
		Where("event_complete = ?", ns.EventComplete).
		Where("event_approving = ?", ns.EventApproving).
		Where("event_running = ?", ns.EventRunning)
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
		switch notification.NotificationType {
		case models.NotificationTypeDingTalk:
			ns.SendDingTalkMessage(notification, "")
		case models.NotificationTypeWebhook:
			ns.SendWebhookMessage(notification, "")
		case models.NotificationTypeWeChat:
			ns.SendWechatMessage(notification, "")
		case models.NotificationTypeSlack:
			ns.SendSlackMessage(notification, "")
		case models.NotificationTypeEmail:
			ns.SendEmailMessage(notification, "")
		}
	}

}

func (ns *NotificationService) SendDingTalkMessage(n models.Notification, message string) {
	dingTalk := NewDingTalkRobot(n.Url, n.Secret)
	if err := dingTalk.SendMarkdownMessage("", message, nil, false); err != nil {

	}
}

func (ns *NotificationService) SendWechatMessage(n models.Notification, message string) {
	wechat := WeChatRobot{Url: n.Url}
	if _, err := wechat.SendMarkdown(message); err != nil {

	}
}

func (ns *NotificationService) SendWebhookMessage(n models.Notification, message string) {
	w := Webhook{Url: n.Url}
	if err := w.Send(message); err != nil {
	}
}

func (ns *NotificationService) SendSlackMessage(n models.Notification, message string) {
	if errs := SendSlack(n.Url, Payload{Text: message, Markdown: true}); len(errs) != 0 {

	}

}

func (ns *NotificationService) SendEmailMessage(n models.Notification, message string) {
	// 获取用户邮箱列表
	users := services.GetUsersByUserIds(db.Get(), n.UserIds)
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
