// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package forms

import "cloudiac/portal/models"

type BatchUpdateVariableForm struct {
	BaseForm
	TplId             models.Id  `json:"tplId" form:"tplId" binding:"omitempty,startswith=tpl-,max=32"` // 模板id
	EnvId             models.Id  `json:"envId" form:"envId" binding:"omitempty,startswith=env-,max=32"` // 环境id
	Variables         []Variable `json:"variables" form:"variables" binding:"omitempty,dive,required"`
	DeleteVariablesId []string   `json:"deleteVariablesId" form:"deleteVariablesId" binding:"omitempty,dive,required,startswith=var-,max=32"` //变量id
}

type UpdateObjectVarsForm struct {
	BaseForm

	Scope    string    `json:"scope" form:"scope" binding:"required,oneof=org template project env" swaggerignore:"true"` // 变量作用域, enum:('org','template','project','env')
	ObjectId models.Id `json:"objectId" binding:"required" swaggerignore:"true"`                                          // 变量所属实例 id

	Variables []Variable `json:"variables" form:"variables" binding:"required,dive,required"` // 变量列表
}

type Variable struct {
	Id          models.Id       `json:"id" form:"id" binding:"omitempty,startswith=var-,max=32"`
	Scope       string          `json:"scope" form:"scope" binding:"omitempty,oneof=org template project env"`    // 应用范围 ('org','template','project','env')
	Type        string          `json:"type" form:"type" binding:"omitempty,oneof=environment terraform ansible"` // 类型 ('environment','terraform','ansible')
	Name        string          `json:"name" form:"name" binding:"required,gte=2,lte=64"`                         // 名称
	Value       string          `json:"value" form:"value" binding:""`                                            // VALUE
	Sensitive   bool            `json:"sensitive" form:"sensitive" `                                              // 是否加密
	Description string          `json:"description" form:"description" binding:"max=255"`                         // 描述
	Options     models.StrSlice `json:"options" form:"options" binding:"omitempty"`                               // 变量下拉列表
}

type SearchVariableForm struct {
	BaseForm
	TplId models.Id `json:"tplId" form:"tplId" binding:"omitempty,startswith=tpl-,max=32"`        // 模板id
	EnvId models.Id `json:"envId" form:"envId" binding:"omitempty,startswith=env-,max=32"`        // 环境id
	Scope string    `json:"scope" form:"scope" binding:"required,oneof=org template project env"` // 应用范围
}
