// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package forms

import "cloudiac/portal/models"

type SearchVariableGroupForm struct {
	NoPageSizeForm
	Q string `form:"q" json:"q" binding:""`
}

type CreateVariableGroupForm struct {
	BaseForm
	Name      string                    `json:"name" form:"name" binding:"required,gte=2,lte=64"`
	Type      string                    `json:"type" form:"type" binding:"required,oneof=environment terraform"`
	Variables []models.VarGroupVariable `json:"variables" form:"variables" binding:"required,dive,required"`
}

type UpdateVariableGroupForm struct {
	BaseForm
	Id        models.Id                 `uri:"id" binding:"required,startswith=vg-,max=32" swaggerignore:"true"`
	Name      string                    `json:"name" form:"name" binding:"omitempty,gte=2,lte=64"`
	Variables []models.VarGroupVariable `json:"variables" form:"variables" binding:"required"`
}

type DeleteVariableGroupForm struct {
	BaseForm
	Id models.Id `uri:"id" binding:"required,startswith=vg-,max=32"`
}

type DetailVariableGroupForm struct {
	BaseForm
	Id models.Id `uri:"id" binding:"required,startswith=vg-,max=32"`
}

type SearchRelationshipForm struct {
	BaseForm
	ObjectType string    `json:"objectType" form:"objectType" binding:"required,oneof=org template project env"` //enum('org','template','project','env')
	TplId      models.Id `json:"tplId" form:"tplId" binding:"omitempty,startswith=tpl-,max=32"`                  // 模板id
	EnvId      models.Id `json:"envId" form:"envId" binding:"omitempty,startswith=env-,max=32"`
}

type BatchUpdateRelationshipForm struct {
	BaseForm
	VarGroupIds    []models.Id `json:"varGroupIds" form:"varGroupIds" binding:"required,dive,required,startswith=vg-,max=32"`
	DelVarGroupIds []models.Id `json:"delVarGroupIds" form:"delVarGroupIds" binding:"omitempty,dive,required,startswith=vg-,max=32"`
	ObjectType     string      `json:"objectType" form:"objectType" binding:"required,oneof=org template project env"` //enum('org','template','project','env')
	ObjectId       models.Id   `json:"objectId" form:"objectId" binding:"required" `
}

type DeleteRelationshipForm struct {
	BaseForm
	Id models.Id `uri:"id" json:"id" form:"id" binding:"required,startswith=vg-,max=32" `
}
