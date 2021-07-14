package forms

import "cloudiac/portal/models"

type UserAuthorization struct {
	UserId models.Id `json:"userId" form:"userId" `                                  // 用户id
	Role   string    `json:"role" form:"role" enums:"'owner,manager,operator,guest"` // 角色 ('owner','manager','operator','guest')
}

type CreateProjectForm struct {
	BaseForm

	Name              string              `json:"name" form:"name" binding:"required"` // 项目名称
	Description       string              `json:"description" form:"description" `     // 项目描述
	UserAuthorization []UserAuthorization `json:"userAuthorization" form:"userAuthorization" `
}

type SearchProjectForm struct {
	PageForm

	Q      string `json:"q" form:"q" `
	Status string `json:"status" form:"status"`
}

type UpdateProjectForm struct {
	BaseForm

	Id                models.Id           `uri:"id" json:"id" swaggerignore:"true"` // 组织ID，swagger 参数通过 param path 指定，这里忽略Id models.Id `uri:"id" json:"id" swaggerignore:"true"` // 组织ID，swagger 参数通过 param path 指定，这里忽略
	Status            string              `json:"status" form:"status" `            // 项目状态 ('enable','disable')
	Name              string              `json:"name" form:"name"`                 // 项目名称
	Description       string              `json:"description" form:"description" `  // 项目描述
	UserAuthorization []UserAuthorization `json:"userAuthorization" form:"userAuthorization" `
}

type DeleteProjectForm struct {
	BaseForm

	Id models.Id `uri:"id" json:"id" swaggerignore:"true"` // 组织ID，swagger 参数通过 param path 指定，这里忽略
}

type DetailProjectForm struct {
	BaseForm

	Id models.Id `uri:"id" json:"id" swaggerignore:"true"` // 组织ID，swagger 参数通过 param path 指定，这里忽略Id models.Id `uri:"id" json:"id" swaggerignore:"true"` // 组织ID，swagger 参数通过 param path 指定，这里忽略
}
