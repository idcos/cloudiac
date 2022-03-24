// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package forms

import "cloudiac/portal/models"

type SearchVariableGroupForm struct {
	NoPageSizeForm
	Q string `form:"q" json:"q" binding:""`
}

type CreateVariableGroupForm struct {
	BaseForm
	Name      string                    `json:"name" form:"name"`
	Type      string                    `json:"type" form:"type"`
	Variables []models.VarGroupVariable `json:"variables" form:"variables" `
}

type UpdateVariableGroupForm struct {
	BaseForm

	Id        models.Id                 `uri:"id" swaggerignore:"true"`
	Name      string                    `json:"name" form:"name"`
	Variables []models.VarGroupVariable `json:"variables" form:"variables" `
}

type DeleteVariableGroupForm struct {
	BaseForm
	Id models.Id `uri:"id"`
}

type DetailVariableGroupForm struct {
	BaseForm
	Id models.Id `uri:"id"`
}

type SearchRelationshipForm struct {
	BaseForm
	ObjectType string    `json:"objectType" form:"objectType" ` //enum('org','template','project','env')
	TplId      models.Id `json:"tplId" form:"tplId" `           // 模板id
	EnvId      models.Id `json:"envId" form:"envId" `
}

type BatchUpdateRelationshipForm struct {
	BaseForm
	VarGroupIds    []models.Id `json:"varGroupIds" form:"varGroupIds" `
	DelVarGroupIds []models.Id `json:"delVarGroupIds" form:"delVarGroupIds" `
	ObjectType     string      `json:"objectType" form:"objectType" ` //enum('org','template','project','env')
	ObjectId       models.Id   `json:"objectId" form:"objectId" `
}

type DeleteRelationshipForm struct {
	BaseForm
	Id models.Id `uri:"id" json:"id" form:"id" `
}
