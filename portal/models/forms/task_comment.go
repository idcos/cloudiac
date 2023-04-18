// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package forms

import "cloudiac/portal/models"

type CreateTaskCommentForm struct {
	BaseForm
	Id      models.Id `uri:"id" json:"id" form:"id" binding:"required,startswith=run-,max=32" swaggerignore:"true"`
	Comment string    `json:"comment" form:"comment" binding:"required,gte=2,lte=255"`
}

type SearchTaskCommentForm struct {
	PageForm
	Id models.Id `uri:"id" json:"id" form:"id" binding:"required,startswith=run-,max=32" swaggerignore:"true"`
}
