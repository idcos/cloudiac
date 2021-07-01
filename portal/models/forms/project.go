package forms

import "cloudiac/portal/models"

type UserAuthorization struct {
	UserId models.Id `json:"userId" form:"userId" `
	Role   string `json:"role" form:"role" `
}

type CreateProjectForm struct {
	BaseForm

	Name              string              `json:"name" form:"name" binding:"required"`
	Description       string              `json:"description" form:"description" `
	UserAuthorization []UserAuthorization `json:"userAuthorization" form:"userAuthorization" `
}

type SearchProjectForm struct {
	BaseForm

	Q string `json:"q" form:"q" `
}

type UpdateProjectForm struct {
	BaseForm

	Id                string              `json:"id" form:"id" binding:"required"`
	Name              string              `json:"name" form:"name"`
	Description       string              `json:"description" form:"description" `
	UserAuthorization []UserAuthorization `json:"userAuthorization" form:"userAuthorization" `
}

type DeleteProjectForm struct {
	BaseForm

	Id models.Id `json:"id" form:"id" binding:"required"`
}

type DetailProjectForm struct {
	BaseForm

	Id models.Id `json:"id" form:"id" binding:"required"`
}
