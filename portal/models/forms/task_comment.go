// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package forms

import "cloudiac/portal/models"

type CreateTaskCommentForm struct {
	BaseForm

	Id      models.Id `uri:"id" json:"id" form:"id" binding:"" swaggerignore:"true"`
	Comment string    `json:"comment" form:"comment" binding:"required"`
}

type SearchTaskCommentForm struct {
	PageForm
	Id models.Id `uri:"id" json:"id" form:"id" swaggerignore:"true"`
}
