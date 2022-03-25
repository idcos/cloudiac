// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package forms

import (
	"cloudiac/portal/models"
)

type CreateOrganizationForm struct {
	BaseForm
	Name        string `form:"name" json:"name" binding:"required,gte=2,lte=64"` // 组织名称
	Description string `form:"description" json:"description" binding:"max=255"` // 组织描述
}

type UpdateOrganizationForm struct {
	BaseForm

	Id models.Id `uri:"id" json:"id" swaggerignore:"true" binding:"required,startswith=org-,max=32"` // 组织ID，swagger 参数通过 param path 指定，这里忽略

	Name        string `form:"name" json:"name" binding:"omitempty,gte=2,lte=64"`                                    // 组织名称
	Description string `form:"description" json:"description" binding:"max=255"`                                     // 组织描述
	RunnerId    string `form:"runnerId" json:"runnerId" binding:"max=255"`                                           // 组织默认部署通道
	Status      string `form:"status" json:"status" binding:"omitempty,oneof=enable disable" enums:"enable,disable"` // 组织状态
}

type SearchOrganizationForm struct {
	NoPageSizeForm

	Q      string `form:"q" json:"q" binding:""`                                                                // 组织名称，支持模糊查询
	Status string `form:"status" json:"status" binding:"omitempty,oneof=enable disable" enums:"enable,disable"` // 组织状态
}

type DeleteOrganizationForm struct {
	BaseForm

	Id models.Id `uri:"id" json:"id" binding:"required,startswith=org-,max=32" swaggerignore:"true"` // 组织ID，swagger 参数通过 param path 指定，这里忽略
}

type DisableOrganizationForm struct {
	BaseForm

	Id models.Id `uri:"id" json:"id" binding:"required,startswith=org-,max=32" swaggerignore:"true"` // 组织ID，swagger 参数通过 param path 指定，这里忽略

	Status string `form:"status" json:"status" binding:"required,oneof=enable disable" enums:"enable,disable"` // 组织状态
}

type DetailOrganizationForm struct {
	BaseForm

	Id models.Id `uri:"id" json:"id" binding:"required,startswith=org-,max=32" swaggerignore:"true"` // 组织ID，swagger 参数通过 param path 指定，这里忽略
}

type OrganizationParam struct {
	BaseForm

	Id models.Id `uri:"id" json:"id" binding:"required,startswith=org-,max=32" swaggerignore:"true"` // 组织ID，swagger 参数通过 param path 指定，这里忽略
}

type InviteUserForm struct {
	BaseForm

	Id     models.Id `uri:"id" json:"id" binding:"required,startswith=org-,max=32" swaggerignore:"true"`                                       // 组织ID
	UserId models.Id `form:"userId"  json:"userId" binding:"required_without_all=Name Email,omitempty,startswith=u-,max=32"`                   // 用户ID，用户ID 或 用户名+邮箱必须填写一个
	Name   string    `form:"name" json:"name" binding:"required_without=UserId,required_with=Email,omitempty,gte=2,lte=32"`                    // 用户名
	Email  string    `form:"email" json:"email" binding:"required_without=UserId,required_with=Name,omitempty,email,max=64"`                   // 电子邮件地址
	Role   string    `form:"role" json:"role" binding:"omitempty,oneof=admin member complianceManager" enums:"admin,member,complianceManager"` // 受邀请用户在组织中的角色，组织管理员：admin，普通用户：member
	Phone  string    `form:"phone" json:"phone" binding:"max=11"`                                                                              // 用户手机号
}

type SearchOrgResourceForm struct {
	PageForm
	Q string `form:"q" json:"q" binding:""` // 资源名称，支持模糊查询
}

type InviteUsersBatchForm struct {
	BaseForm

	Id    models.Id `uri:"id" json:"id" binding:"required,startswith=org-,max=32" swaggerignore:"true"`   // 组织ID
	Email []string  `form:"email" json:"email" binding:"required,dive,required,email,max=64"`             // 电子邮件地址
	Role  string    `form:"role" json:"role" binding:"omitempty,oneof=admin member" enums:"admin,member"` // 受邀请用户在组织中的角色，组织管理员：admin，普通用户：member
}
