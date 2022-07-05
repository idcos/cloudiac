// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package forms

import (
	"cloudiac/portal/models"
)

type CreateTokenForm struct {
	BaseForm

	Type        string    `json:"type" form:"type" binding:"required,max=255"`                       //类型
	Role        string    `json:"role" form:"role" binding:"max=255"`                                // token角色
	ExpiredAt   string    `json:"expiredAt" form:"expiredAt" `                                       // 过期时间
	Description string    `json:"description" form:"description" binding:"max=255" `                 //描述
	EnvId       models.Id `json:"envId" form:"envId" binding:"omitempty,startswith=env-,max=32"`     //创建触发器token时必传，其他可不传
	Action      string    `json:"action" form:"action" binding:"omitempty,oneof=plan apply destroy"` //创建触发器token时必传，其他可不传('apply','plan','destroy')
}

type UpdateTokenForm struct {
	BaseForm
	Id          models.Id `uri:"id" form:"id" json:"id" binding:"required,startswith=t-,max=32" swaggerignore:"true"`
	Status      string    `form:"status" json:"status" binding:"omitempty,oneof=enable disable"`
	Description string    `json:"description" form:"description" binding:"max=255" ` //描述
}

type SearchTokenForm struct {
	PageForm
	Q      string `form:"q" json:"q" binding:""`                                         //"模糊搜索"
	Status string `form:"status" json:"status" binding:"omitempty,oneof=enable disable"` //"ApiToken状态"
}

type DeleteTokenForm struct {
	BaseForm
	Id models.Id `uri:"id" form:"id" json:"id" binding:"required,startswith=t-,max=32"`
}

type VcsWebhookUrlForm struct {
	BaseForm
	EnvId models.Id `json:"envId" form:"envId" binding:"required,startswith=env-,max=32"`
	//Action string    `json:"action" form:"action" binding:"required"`
}

type LoginForm struct {
	BaseForm

	Email    string `json:"email" form:"email" binding:"required,email,max=64"`        // 登陆的用户电子邮箱地址
	Password string `json:"password" form:"password" binding:"required,ascii,max=255"` // 密码
}

type EmailForm struct {
	BaseForm

	Email string `json:"email" form:"email" binding:"required,email,max=64"` // 登陆的用户电子邮箱地址
}

type PasswordResetEmailForm struct {
	BaseForm

	Email string `json:"email" form:"email" binding:"required,email,max=64"` // 登陆的用户电子邮箱地址
}

type PasswordResetForm struct {
	BaseForm

	Password string `json:"password" form:"password" binding:"required,ascii,max=30,min=6"` // 密码
}

type ApiTriggerHandler struct {
	BaseForm
	Token string `json:"token" form:"token" binding:"required,max=255"`
}

type RegistryForm struct {
	BaseForm

	Name     string `json:"name" form:"name" binding:"required,max=64"`
	Email    string `json:"email" form:"email" binding:"required,email,max=64"`        // 登陆的用户电子邮箱地址
	Password string `json:"password" form:"password" binding:"required,ascii,max=255"` // 密码

	Phone   string `json:"phone" form:"phone" binding:""`
	Company string `json:"company" form:"company" binding:""`
}
