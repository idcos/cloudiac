// Copyright 2021 CloudJ Company Limited. All rights reserved.

package forms

import "cloudiac/portal/models"

type UserAuthorization struct {
	UserId models.Id `json:"userId" form:"userId" `                                     // 用户id
	Role   string    `json:"role" form:"role" enums:"'manager,approver,operator,guest"` // 角色 ('owner','manager','operator','guest')
}

type CreateProjectForm struct {
	BaseForm

	Name              string              `json:"name" form:"name" binding:"required"` // 项目名称
	Description       string              `json:"description" form:"description" `     // 项目描述
	UserAuthorization []UserAuthorization `json:"userAuthorization" form:"userAuthorization" `
}

type SearchProjectForm struct {
	NoPageSizeForm

	Q      string `json:"q" form:"q" `
	Status string `json:"status" form:"status"`
}

type UpdateProjectForm struct {
	BaseForm

	Id          models.Id `uri:"id" json:"id" swaggerignore:"true"`
	Status      string    `json:"status" form:"status" `           // 项目状态 ('enable','disable')
	Name        string    `json:"name" form:"name"`                // 项目名称
	Description string    `json:"description" form:"description" ` // 项目描述
}

type DeleteProjectForm struct {
	BaseForm

	Id models.Id `uri:"id" json:"id" swaggerignore:"true"`
}

type DetailProjectForm struct {
	BaseForm

	Id models.Id `uri:"id" json:"id" swaggerignore:"true"`
}
