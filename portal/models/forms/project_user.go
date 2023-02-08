// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package forms

import "cloudiac/portal/models"

type CreateProjectUserForm struct {
	BaseForm

	UserId []models.Id `json:"userId" form:"userId" binding:"required,dive,required,startswith=u-,max=32"`                                         // 用户id
	Role   string      `json:"role" form:"role" binding:"required,oneof=manager approver operator guest" enums:"'manager,approver,operator,guest"` // 角色 (manager,approver,operator,guest)
}

type DeleteProjectOrgUserForm struct {
	BaseForm
	Id string `uri:"id" json:"id" form:"id" binding:"required,startswith=u-,max=32"`
}

type UpdateProjectUserForm struct {
	BaseForm
	Id models.Id `uri:"id" json:"id" form:"id" binding:"required,startswith=u-,max=32" swaggerignore:"true"`
	//UserId models.Id `json:"userId" form:"userId" `                                     // 用户id
	Role string `json:"role" form:"role" binding:"required,oneof=manager approver operator guest" enums:"'manager,approver,operator,guest"` // 角色 (manager,approver,operator,guest)
}

type SearchProjectAuthorizationUserForm struct {
	PageForm
}
