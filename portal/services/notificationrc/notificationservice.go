package notificationrc

import (
	"cloudiac/configs"
	"cloudiac/portal/consts"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/utils"
	"cloudiac/utils/logs"
	"cloudiac/utils/mail"
	"fmt"
)

type NotificationService struct {
	Tpl            *models.Template     `json:"tpl" form:"tpl" `
	Project        *models.Project      `json:"project" form:"project" `
	Org            *models.Organization `json:"org" form:"org" `
	OrgId          models.Id            `json:"orgId" form:"orgId" `
	ProjectId      models.Id            `json:"projectId" form:"projectId" `
	Env            *models.Env          `json:"env" form:"env" `
	Task           *models.Task         `json:"task" form:"task" `
	EventFailed    bool                 `json:"eventFailed" form:"eventFailed" `
	EventComplete  bool                 `json:"eventComplete" form:"eventComplete" `
	EventApproving bool                 `json:"eventApproving" form:"eventApproving" `
	EventRunning   bool                 `json:"eventRunning" form:"eventRunning" `
}

type NotificationOptions struct {
	Tpl            *models.Template     `json:"tpl" form:"tpl" `
	Project        *models.Project      `json:"project" form:"project" `
	Org            *models.Organization `json:"org" form:"org" `
	OrgId          models.Id            `json:"orgId" form:"orgId" `
	ProjectId      models.Id            `json:"projectId" form:"projectId" `
	Env            *models.Env          `json:"env" form:"env" `
	Task           *models.Task         `json:"task" form:"task" `
	EventFailed    bool                 `json:"eventFailed" form:"eventFailed" `
	EventComplete  bool                 `json:"eventComplete" form:"eventComplete" `
	EventApproving bool                 `json:"eventApproving" form:"eventApproving" `
	EventRunning   bool                 `json:"eventRunning" form:"eventRunning" `
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
		Tpl:            options.Tpl,
		Project:        options.Project,
		Org:            options.Org,
	}
}

func (ns *NotificationService) SendMessage() {
	notifications, tplNotificationTemplate, markdownNotificationTemplate := ns.FindNotificationsAndMessageTpl()
	u := models.User{}
	_ = db.Get().Where("id = ?", ns.Task.CreatorId).First(&u)

	markdownNotificationTemplate = utils.SprintTemplate(markdownNotificationTemplate, struct {
		Creator      string
		OrgName      string
		ProjectName  string
		TemplateName string
		Revision     string
		EnvName      string
		Addr         string
		ResAdded     *int
		ResChanged   *int
		ResDestroyed *int
		Message      string
	}{
		Creator:      u.Name,
		OrgName:      ns.Org.Name,
		ProjectName:  ns.Project.Name,
		TemplateName: ns.Tpl.Name,
		Revision:     ns.Tpl.RepoRevision,
		EnvName:      ns.Env.Name,
		//http://10.0.2.135/org/org-c3vm0ljn6m8705n103ug/project/p-c3vm7tbn6m80l6s918m0/m-project-env/detail/env-c4adojrn6m83bocqu9h0/deployHistory/task/run-c4adojrn6m83bocqu9hg
		Addr:         fmt.Sprintf("%s/org/%s/project/%s/m-project-env/detail/%s/deployHistory/task/%s", configs.Get().Portal.Address, ns.Org.Id, ns.ProjectId, ns.Env.Id, ns.Task.Id),
		ResAdded:     ns.Task.Result.ResAdded,
		ResChanged:   ns.Task.Result.ResChanged,
		ResDestroyed: ns.Task.Result.ResDestroyed,
		Message:      ns.Task.Message,
	})
	tplNotificationTemplate = utils.SprintTemplate(tplNotificationTemplate, NotificationOptions{Env: ns.Env, Task: ns.Task})
	for _, notification := range notifications {
		//todo 消息通知模板
		switch notification.NotificationType {
		case models.NotificationTypeDingTalk:
			ns.SendDingTalkMessage(notification, markdownNotificationTemplate)
		case models.NotificationTypeWebhook:
			ns.SendWebhookMessage(notification, markdownNotificationTemplate)
		case models.NotificationTypeWeChat:
			ns.SendWechatMessage(notification, markdownNotificationTemplate)
		case models.NotificationTypeSlack:
			ns.SendSlackMessage(notification, markdownNotificationTemplate)
		case models.NotificationTypeEmail:
			ns.SendEmailMessage(notification, tplNotificationTemplate)
		}
	}

}

func (ns *NotificationService) SendDingTalkMessage(n models.Notification, message string) {
	dingTalk := NewDingTalkRobot(n.Url, n.Secret)
	if err := dingTalk.SendMarkdownMessage("CloudIaC平台系统通知", message, nil, false); err != nil {
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
	if err := mail.SendMail(emails, "CloudIaC平台系统通知", message); err != nil {

	}
}

func (ns *NotificationService) FindNotificationsAndMessageTpl() ([]models.Notification, string, string) {
	orgNotification := make([]models.Notification, 0)
	projectNotification := make([]models.Notification, 0)
	notifications := make([]models.Notification, 0)
	dbSess := db.Get().Debug().Where("org_id = ?", ns.OrgId)
	var (
		tplNotificationTemplate      string
		markdownNotificationTemplate string
	)
	if ns.EventFailed {
		dbSess = dbSess.Where("event_failed = ?", ns.EventFailed)
		tplNotificationTemplate = consts.IacTaskFailedTpl
		markdownNotificationTemplate = consts.IacTaskFailedMarkdown
	}
	if ns.EventComplete {
		dbSess = dbSess.Where("event_complete = ?", ns.EventComplete)
		tplNotificationTemplate = consts.IacTaskCompleteTpl
		markdownNotificationTemplate = consts.IacTaskCompleteMarkdown
	}
	if ns.EventApproving {
		dbSess = dbSess.Where("event_approving = ?", ns.EventApproving)
		tplNotificationTemplate = consts.IacTaskApprovingTpl
		markdownNotificationTemplate = consts.IacTaskApprovingMarkdown
	}
	if ns.EventRunning {
		dbSess = dbSess.Where("event_running = ?", ns.EventRunning)
		tplNotificationTemplate = consts.IacTaskRunning
		markdownNotificationTemplate = consts.IacTaskRunningMarkdown
	}
	// 查询需要组织下需要通知的人
	if err := dbSess.
		Where("project_id = '' or project_id = null").
		Find(&orgNotification); err != nil {
		return notifications, tplNotificationTemplate, markdownNotificationTemplate
	}
	// 查询需要项目下需要通知的人
	if err := dbSess.
		Where("project_id = ?", ns.ProjectId).
		Find(&projectNotification); err != nil {
		return notifications, tplNotificationTemplate, markdownNotificationTemplate
	}

	// 将需要通知的数据进行整理
	notifications = append(notifications, orgNotification...)
	notifications = append(notifications, projectNotification...)
	return notifications, tplNotificationTemplate, markdownNotificationTemplate
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
