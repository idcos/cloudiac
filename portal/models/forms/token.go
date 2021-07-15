package forms

import (
	"cloudiac/portal/models"
	"cloudiac/utils"
)

type CreateTokenForm struct {
	PageForm

	Type        string         `json:"type" form:"type" `               //类型
	Role        string         `json:"role" form:"role" `               // token角色
	ExpiredAt   utils.JSONTime `json:"expiredAt" form:"expiredAt" `     // 过期时间
	Description string         `json:"description" form:"description" ` //描述
}

type UpdateTokenForm struct {
	PageForm
	Id          models.Id `uri:"id" form:"id" json:"id" binding:"required"`
	Status      string    `form:"status" json:"status" binding:"required"`
	Description string    `json:"description" form:"description" ` //描述
}

type SearchTokenForm struct {
	PageForm
	Q      string `form:"q" json:"q" binding:""`
	Status string `form:"status" json:"status"`
}

type DeleteTokenForm struct {
	PageForm
	Id models.Id `uri:"id" form:"id" json:"id" binding:"required"`
}

type LoginForm struct {
	BaseForm

	Email    string `json:"email" form:"email" binding:"required,email"` // 登陆的用户电子邮箱地址
	Password string `json:"password" form:"password" binding:"required"` // 密码
}
