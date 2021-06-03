package forms

type CfgInfo struct {
	EmailAddress string `form:"emailAddress" json:"emailAddress"`
	UserId       int    `form:"userId" json:"userId"`
	WebUrl       string `form:"webUrl" json:"webUrl"`
	UserName     string `form:"userName" json:"userName"`
}

type UpdateNotificationCfgForm struct {
	BaseForm
	NotificationId   int     `form:"notificationId" json:"notificationId" binding:"required"`
	NotificationType string  `form:"notificationType" json:"notificationType" binding:"required"`
	EventType        string  `form:"eventType" json:"eventType" binding:"required"`
	CfgInfo          CfgInfo `form:"cfgInfo" json:"cfgInfo"`
}

type CreateNotificationCfgForm struct {
	BaseForm
	NotificationType string  `form:"notificationType" json:"notificationType" binding:"required"`
	EventType        string  `form:"eventType" json:"eventType" binding:"required"`
	UserIds          []uint  `form:"userIds" json:"userIds"`
	CfgInfo          CfgInfo `form:"cfgInfo" json:"cfgInfo"`
}

type DeleteNotificationCfgForm struct {
	BaseForm
	Id uint `form:"id" json:"id" binding:"required"`
}

type CreateOrganizationForm struct {
	BaseForm
	Name                   string `form:"name" json:"name" binding:"required,gte=2,lte=32"`
	Description            string `form:"description" json:"description" binding:""`
	VcsType                string `form:"vcsType" json:"vcsType" binding:""`
	VcsVersion             string `form:"vcsVersion" json:"vcsVersion" binding:""`
	VcsAuthInfo            string `form:"vcsAuthInfo" json:"vcsAuthInfo" binding:""`
	DefaultRunnerAddr      string `json:"defaultRunnerAddr" `
	DefaultRunnerPort      uint   `json:"defaultRunnerPort" `
	DefaultRunnerServiceId string `json:"defaultRunnerServiceId"`
}

type UpdateOrganizationForm struct {
	BaseForm
	Id                     uint   `form:"id" json:"id" binding:""`
	Name                   string `form:"name" json:"name" binding:""`
	Description            string `form:"description" json:"description" binding:"max=255"`
	VcsType                string `form:"vcsType" json:"vcsType" binding:""`
	VcsVersion             string `form:"vcsVersion" json:"vcsVersion" binding:""`
	VcsAuthInfo            string `form:"vcsAuthInfo" json:"vcsAuthInfo" binding:""`
	DefaultRunnerAddr      string `json:"defaultRunnerAddr" `
	DefaultRunnerPort      uint   `json:"defaultRunnerPort" `
	DefaultRunnerServiceId string `json:"defaultRunnerServiceId"`
}

type SearchOrganizationForm struct {
	BaseForm

	Q      string `form:"q" json:"q" binding:""`
	Status string `form:"status" json:"status"`
}

type DeleteOrganizationForm struct {
	BaseForm
	Id uint `form:"id" json:"id" binding:"required"`
}

type DisableOrganizationForm struct {
	BaseForm

	Id     uint   `form:"id" json:"id" binding:"required"`
	Status string `form:"status" json:"status" binding:"required"`
}

type DetailOrganizationForm struct {
	BaseForm
	Id uint `form:"id" json:"id" binding:"required"`
}
