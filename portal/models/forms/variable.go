package forms

import "cloudiac/portal/models"

type CreateVariableForm struct {
	BaseForm
	TplId           models.Id   `json:"tplId" form:"tplId" ` // 模板id
	EnvId           models.Id   `json:"envId" form:"envId" ` // 环境id
	Variables       []Variables `json:"variables" form:"variables" `
	DeleteVariables []string    `json:"deleteVariables" form:"deleteVariables" `
}

type Variables struct {
	Id          models.Id `json:"id" form:"id" `
	Scope       string    `json:"scope" form:"scope" `             // 应用范围
	Type        string    `json:"type" form:"type" `               // 类型
	Name        string    `json:"name" form:"name" `               // 名称
	Value       string    `json:"value" form:"value" `             // VALUE
	Sensitive   bool      `json:"sensitive" form:"sensitive" `     // 是否加密
	Description string    `json:"description" form:"description" ` // 描述
}

type SearchVariableForm struct {
	PageForm
	TplId models.Id `json:"tplId" form:"tplId" ` // 模板id
	EnvId models.Id `json:"envId" form:"envId" ` // 环境id
	Scope string    `json:"scope" form:"scope" ` // 应用范围
}
