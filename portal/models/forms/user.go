// Copyright 2021 CloudJ Company Limited. All rights reserved.

package forms

import "cloudiac/portal/models"

type CreateUserForm struct {
	BaseForm

	Name  string `form:"name" json:"name" binding:"required,gte=2,lte=32"` // 用户名
	Phone string `form:"phone" json:"phone" binding:"max=11"`              // 电话
	Email string `form:"email" json:"email" binding:"required,email"`      // 电子邮件地址
}

type UpdateUserForm struct {
	BaseForm

	Id          models.Id         `uri:"id" json:"id" binding:"" swaggerignore:"true"`         // 用户ID
	Name        string            `form:"name" json:"name" binding:""`                         // 用户名
	Phone       string            `form:"phone" json:"phone" binding:""`                       // 电话
	OldPassword string            `form:"oldPassword" json:"oldPassword" binding:""`           // 原始密码
	NewPassword string            `form:"newPassword" json:"newPassword" binding:""`           // 新密码
	NewbieGuide map[string]uint64 `form:"newbieGuide" json:"newbieGuide" swaggertype:"string"` // 新用户向导内容
}

type SearchUserForm struct {
	PageForm

	Q       string `form:"q" json:"q" binding:""`                       // 用户名，支持模糊查询
	Status  string `form:"status" json:"status" enums:"enable,disable"` // 状态
	Exclude string `form:"exclude" json:"exclude" enums:"org,project"`  // 排除用户方式：1. org：排除当前组织用户 2. project: 排除当前项目用户
}

type DeleteUserForm struct {
	BaseForm

	Id models.Id `uri:"id" json:"id" binding:"" swaggerignore:"true"` // 用户ID
}

type DisableUserForm struct {
	BaseForm

	Id     models.Id `uri:"id" json:"id" binding:"" swaggerignore:"true"`                    // 用户ID
	Status string    `form:"status" json:"status" binding:"required" enums:"enable,disable"` // 状态
}

type DetailUserForm struct {
	BaseForm

	Id models.Id `uri:"id" json:"id" binding:"" swaggerignore:"true"` // 用户ID
}

type AddUserOrgRelForm struct {
	BaseForm

	Id     models.Id `uri:"id" json:"id" binding:"" swaggerignore:"true"`      // 组织ID
	UserId models.Id `form:"userId" json:"userId" binding:""`                  // 用户ID
	Role   string    `form:"role" json:"role" binding:"" enums:"admin,member"` // 用户在组织中的角色，组织管理员：admin，普通用户：member，默认 member
}

type DeleteUserOrgRelForm struct {
	BaseForm

	Id     models.Id `uri:"id" json:"id" binding:"" swaggerignore:"true"`         // 组织ID
	UserId models.Id `uri:"userId" json:"userId" binding:"" swaggerignore:"true"` // 用户ID
}

type UpdateUserOrgRelForm struct {
	BaseForm

	Id     models.Id `uri:"id" json:"id" binding:"" swaggerignore:"true"`              // 组织ID
	UserId models.Id `uri:"userId" json:"userId" binding:"" swaggerignore:"true"`      // 用户ID
	Role   string    `form:"role" json:"role" binding:"required" enums:"admin,member"` // 用户在组织中的角色，组织管理员：admin，普通用户：member，默认 member
}
