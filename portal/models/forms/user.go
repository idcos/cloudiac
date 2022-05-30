// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package forms

import "cloudiac/portal/models"

type CreateUserForm struct {
	BaseForm

	Name   string `form:"name" json:"name" binding:"required,gte=2,lte=32"`   // 用户名
	Phone  string `form:"phone" json:"phone" binding:"max=11"`                // 电话
	Email  string `form:"email" json:"email" binding:"required,email,max=64"` // 电子邮件地址
	IsLdap bool   `form:"isLdap" json:"isLdap" default:"false"`
}

type UpdateUserForm struct {
	BaseForm

	Id          models.Id         `uri:"id" json:"id" binding:"required,startswith=u-,max=32" swaggerignore:"true"`     // 用户ID
	Name        string            `form:"name" json:"name" binding:"omitempty,gte=2,lte=64"`                            // 用户名
	Phone       string            `form:"phone" json:"phone" binding:"max=11"`                                          // 电话
	OldPassword string            `form:"oldPassword" json:"oldPassword" binding:"omitempty,ascii"`                     // 原始密码
	NewPassword string            `form:"newPassword" json:"newPassword" binding:"omitempty,ascii,nefield=OldPassword"` // 新密码
	NewbieGuide map[string]uint64 `form:"newbieGuide" json:"newbieGuide" binding:"" swaggertype:"string"`               // 新用户向导内容
}

type SearchUserForm struct {
	NoPageSizeForm

	Q       string `form:"q" json:"q" binding:""`                                                                // 用户名，支持模糊查询
	Status  string `form:"status" json:"status" binding:"omitempty,oneof=enable disable" enums:"enable,disable"` // 状态
	Exclude string `form:"exclude" json:"exclude" binding:"omitempty,oneof=org project" enums:"org,project"`     // 排除用户方式：1. org：排除当前组织用户 2. project: 排除当前项目用户
}

type DeleteUserForm struct {
	BaseForm

	Id models.Id `uri:"id" json:"id" binding:"required,startswith=u-,max=32" swaggerignore:"true"` // 用户ID
}

type DisableUserForm struct {
	BaseForm

	Id     models.Id `uri:"id" json:"id" binding:"required,startswith=u-,max=32" swaggerignore:"true"`            // 用户ID
	Status string    `form:"status" json:"status" binding:"required,oneof=enable disable" enums:"enable,disable"` // 状态
}

type DetailUserForm struct {
	BaseForm
	Id models.Id `uri:"id" json:"id" binding:"required,startswith=u-,max=32" swaggerignore:"true"` // 用户ID
}

type AddUserOrgRelForm struct {
	BaseForm

	Id     models.Id `uri:"id" json:"id" binding:"required,startswith=org-,max=32" swaggerignore:"true"`  // 组织ID
	UserId models.Id `form:"userId" json:"userId" binding:"required,startswith=u-,max=32"`                // 用户ID
	Role   string    `form:"role" json:"role" binding:"required,oneof=admin member" enums:"admin,member"` // 用户在组织中的角色，组织管理员：admin，普通用户：member，默认 member
}

type DeleteUserOrgRelForm struct {
	BaseForm

	Id     models.Id `uri:"id" json:"id" binding:"required,startswith=org-,max=32"  swaggerignore:"true"`       // 组织ID
	UserId models.Id `uri:"userId" json:"userId" binding:"required,startswith=u-,max=32"  swaggerignore:"true"` // 用户ID
}

type UpdateUserOrgRelForm struct {
	BaseForm

	Id     models.Id `uri:"id" json:"id" binding:"required,startswith=org-,max=32" swaggerignore:"true"`                                      // 组织ID
	UserId models.Id `uri:"userId" json:"userId" binding:"required,contains=u-,max=32" swaggerignore:"true"`                                  // 用户ID
	Role   string    `form:"role" json:"role" binding:"required,oneof=admin complianceManager member" enums:"admin,complianceManager,member"` // 用户在组织中的角色，组织管理员：admin，普通用户：member，默认 member
}

type UpdateUserOrgForm struct {
	BaseForm

	Id     models.Id `uri:"id" json:"id" binding:"required,startswith=org-,max=32" swaggerignore:"true"`     // 组织ID
	UserId models.Id `uri:"userId" json:"userId" binding:"required,contains=u-,max=32" swaggerignore:"true"` // 用户ID
	Name   string    `form:"name" json:"name" binding:"gte=2,lte=32"`                                        // 用户名
	Phone  string    `form:"phone" json:"phone" binding:"max=11"`
	Role   string    `form:"role" json:"role" binding:"required,oneof=admin complianceManager member" enums:"admin,complianceManager,member"` // 用户在组织中的角色，组织管理员：admin，普通用户：member，默认 member
}
