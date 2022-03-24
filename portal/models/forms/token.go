// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package forms

import (
	"cloudiac/portal/models"
)

type CreateTokenForm struct {
	BaseForm

	Type        string    `json:"type" form:"type" binding:"required"` //类型
	Role        string    `json:"role" form:"role" `                   // token角色
	ExpiredAt   string    `json:"expiredAt" form:"expiredAt" `         // 过期时间
	Description string    `json:"description" form:"description" `     //描述
	EnvId       models.Id `json:"envId" form:"envId"`                  //创建触发器token时必传，其他可不传
	Action      string    `json:"action" form:"action"`                //创建触发器token时必传，其他可不传('apply','plan','destroy')
}

type UpdateTokenForm struct {
	BaseForm
	Id          models.Id `uri:"id" form:"id" json:"id" binding:"required" swaggerignore:"true"`
	Status      string    `form:"status" json:"status" binding:"required"`
	Description string    `json:"description" form:"description" ` //描述
}

type SearchTokenForm struct {
	PageForm
	Q      string `form:"q" json:"q" binding:""` //"模糊搜索"
	Status string `form:"status" json:"status"`  //"ApiToken状态"
}

type DeleteTokenForm struct {
	BaseForm
	Id models.Id `uri:"id" form:"id" json:"id" binding:"required"`
}

type VcsWebhookUrlForm struct {
	BaseForm
	EnvId models.Id `json:"envId" form:"envId" binding:"required"`
	//Action string    `json:"action" form:"action" binding:"required"`
}

type LoginForm struct {
	BaseForm

	Email    string `json:"email" form:"email" binding:"required,email"` // 登陆的用户电子邮箱地址
	Password string `json:"password" form:"password" binding:"required"` // 密码
}

type ApiTriggerHandler struct {
	BaseForm
	Token string `json:"token" form:"token" binding:"required"`
}
