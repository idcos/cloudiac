// Copyright 2021 CloudJ Company Limited. All rights reserved.

package forms

import "cloudiac/portal/models"

type CreateProjectUserForm struct {
	BaseForm

	UserId models.Id `json:"userId" form:"userId" `                                     // 用户id
	Role   string    `json:"role" form:"role" enums:"'manager,approver,operator,guest"` // 角色 (manager,approver,operator,guest)
}

type DeleteProjectOrgUserForm struct {
	BaseForm
	Id string `uri:"id" json:"id" form:"id" `
}

type UpdateProjectUserForm struct {
	BaseForm
	Id models.Id `uri:"id" json:"id" form:"id" `
	//UserId models.Id `json:"userId" form:"userId" `                                     // 用户id
	Role string `json:"role" form:"role" enums:"'manager,approver,operator,guest"` // 角色 (manager,approver,operator,guest)
}

type SearchProjectAuthorizationUserForm struct {
	PageForm
}
