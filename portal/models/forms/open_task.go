// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package forms

type VarOpen struct {
	//Id          string `form:"id" json:"id" binding:"required"`
	Name  string `json:"name" form:"name"  binding:"required"`
	Value string `form:"value" json:"value" binding:"required"`

	//IsSecret    *bool  `form:"isSecret" json:"isSecret" binding:"required,default:false"`
	//Type        string `form:"type" json:"type" binding:"required,default:env"`
	//Description string `form:"description" json:"description" binding:""`
}

type Runner struct {
	ServiceID string   `json:"serviceId" form:"serviceId" `
	Tags      []string `json:"tags" form:"tags" `
}

type Account struct {
	AccessKeyId     string `json:"accessKeyId" form:"accessKeyId" `
	SecretAccessKey string `json:"secretAccessKey" form:"secretAccessKey" `
}

type CreateTaskOpenForm struct {
	PageForm
	TemplateGuid  string    `json:"templateGuid" form:"templateGuid" binding:"required"`
	Vars          []VarOpen `form:"vars" json:"vars"`
	Account       Account   `json:"account" form:"account" `
	Runner        Runner    `json:"runner" form:"runner" `
	CommitId      string    `json:"commitId" form:"commitId" binding:"required"`
	Source        string    `json:"source" form:"source"`
	TransactionId string    `json:"transactionId"`
}
