// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package forms

import "cloudiac/portal/models"

type UserAuthorization struct {
	UserId models.Id `json:"userId" form:"userId" `                                     // 用户id
	Role   string    `json:"role" form:"role" enums:"'manager,approver,operator,guest"` // 角色 ('owner','manager','operator','guest')
}

type CreateProjectForm struct {
	BaseForm

	Name              string              `json:"name" form:"name" binding:"required,gte=2,lte=64"` // 项目名称
	Description       string              `json:"description" form:"description" binding:"max=255"` // 项目描述
	UserAuthorization []UserAuthorization `json:"userAuthorization" form:"userAuthorization" `
}

type SearchProjectForm struct {
	NoPageSizeForm

	Q         string    `json:"q" form:"q" `
	Status    string    `json:"status" form:"status" binding:"omitempty,oneof=enable disable"` // 过滤项目状态(enable/disable, 默认为 enable)
	WithStat  bool      `json:"withStat" form:"withStat"`
	ProjectId models.Id `json:"projectId" form:"projectId"`
}

type UpdateProjectForm struct {
	BaseForm

	Id          models.Id `uri:"id" json:"id" swaggerignore:"true" binding:"required,startswith=p-,max=32"`
	Status      string    `json:"status" form:"status" binding:"omitempty,oneof=enable disable"` // 项目状态 ('enable','disable')
	Name        string    `json:"name" form:"name" binding:"omitempty,gte=2,lte=64" `            // 项目名称
	Description string    `json:"description" form:"description" binding:"max=255"`              // 项目描述
}

type DeleteProjectForm struct {
	BaseForm

	Id models.Id `uri:"id" json:"id" swaggerignore:"true" binding:"required,startswith=p-,max=32"`
}

type DetailProjectForm struct {
	BaseForm

	Id models.Id `uri:"id" json:"id" binding:"required,startswith=p-,max=32" swaggerignore:"true"`
}

type ProjectStatForm struct {
	BaseForm
	ProjectId models.Id `uri:"id" json:"id" binding:"required"`
	Limit     int       `form:"limit" json:"limit"`
}

type SearchProjectResourceForm struct {
	PageForm
	Q         string `form:"q" json:"q" binding:""`                 // 资源名称，支持模糊查询
	EnvIds    string `form:"envIds" json:"envIds" binding:""`       // 环境id列表
	Providers string `form:"providers" json:"providers" binding:""` // provider 名称列表
}
