package forms

import "cloudiac/portal/models"

type CreateUserForm struct {
	BaseForm

	Name  string `form:"name" json:"name" binding:"required,gte=2,lte=32"`
	Phone string `form:"phone" json:"phone" binding:"max=11"`
	Email string `form:"email" json:"email" binding:"required,email"`
}

type UpdateUserForm struct {
	BaseForm

	Id    models.Id `form:"id" json:"id" binding:""`
	Name  string    `form:"name" json:"name" binding:""`
	Phone string    `form:"phone" json:"phone" binding:""`
	//Email       string `form:"email" json:"email" binding:""`	// 邮箱不可编辑
	OldPassword string            `form:"oldPassword" json:"oldPassword" binding:""`
	NewPassword string            `form:"newPassword" json:"newPassword" binding:""`
	NewbieGuide map[string]uint64 `json:"newbieGuide" form:"newbieGuide" `
}

type SearchUserForm struct {
	PageForm

	Q      string `form:"q" json:"q" binding:""`
	Status string `form:"status" json:"status"`
}

type DeleteUserForm struct {
	BaseForm
	Id models.Id `form:"id" json:"id" binding:"required"`
}

type DisableUserForm struct {
	BaseForm

	Id     models.Id `form:"id" json:"id" binding:"required"`
	Status string    `form:"status" json:"status" binding:"required"`
}

type DetailUserForm struct {
	BaseForm
	Id models.Id `form:"id" json:"id" binding:"required"`
}

type InviteUserForm struct {
	BaseForm

	BaseURL   string      `json:"baseURL" form:"baseURL" binding:"required"`
	Emails    []string    `json:"emails" form:"email" binding:"required"`
	Customers []models.Id `json:"customers" form:"customer" binding:""`
	Roles     []string    `json:"roles" form:"role"`
}

type LoginForm struct {
	BaseForm

	Email    string `json:"email" form:"email" binding:"required"`
	Password string `json:"password" form:"password" binding:"required"`
}

type AdminSearchForm struct {
	PageForm
}
