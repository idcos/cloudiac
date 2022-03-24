// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package forms

import "cloudiac/portal/models"

type Var struct {
	Id          string `form:"id" json:"id" binding:"required"`
	Key         string `form:"key" json:"key" binding:"required"`
	Value       string `form:"value" json:"value" binding:"required"`
	IsSecret    *bool  `form:"isSecret" json:"isSecret" binding:"required,default:false"`
	Type        string `form:"type" json:"type" binding:"required,default:env"`
	Description string `form:"description" json:"description" binding:""`
}

type CreateTemplateForm struct {
	BaseForm

	Name         string      `form:"name" json:"name" binding:"required,gte=2,lte=64"`
	Description  string      `form:"description" json:"description" binding:"max=255"`
	RepoId       string      `form:"repoId" json:"repoId" binding:"required,max=255"`
	RepoFullName string      `form:"repoFullName" json:"repoFullName" binding:"required,max=255"`
	RepoRevision string      `form:"repoRevision" json:"repoRevision" binding:"max=64"`
	Extra        string      `form:"extra" json:"extra"`
	Workdir      string      `form:"workdir" json:"workdir" binding:"max=255"`
	VcsId        models.Id   `form:"vcsId" json:"vcsId" binding:"required,startswith=vcs-,max=32"`
	Playbook     string      `json:"playbook" form:"playbook" binding:"omitempty,endswith=.yml,max=255"`
	PlayVarsFile string      `json:"playVarsFile" form:"playVarsFile" binding:"max=255"`
	TfVarsFile   string      `form:"tfVarsFile" json:"tfVarsFile" binding:"max=255"`
	ProjectId    []models.Id `form:"projectId" json:"projectId" binding:"required,dive,required,startswith=p-,max=32"` // 项目ID
	TfVersion    string      `form:"tfVersion" json:"tfVersion" binding:"max=255"`                                     // 模版使用terraform版本号

	Variables []Variable `json:"variables" form:"variables" binding:"omitempty,dive,required"`

	VarGroupIds    []models.Id `json:"varGroupIds" form:"varGroupIds" binding:"omitempty,dive,required,startswith=vg-,max=32"`
	DelVarGroupIds []models.Id `json:"delVarGroupIds" form:"delVarGroupIds" binding:"omitempty,dive,required,startswith=vg-,max=32"`
	PolicyEnable   bool        `json:"policyEnable" form:"policyEnable"`                                                        // 是否开启合规检测
	PolicyGroup    []models.Id `json:"policyGroup" form:"policyGroup" binding:"omitempty,dive,required,startswith=pog-,max=32"` // 绑定的合规策略组
	TplTriggers    []string    `json:"tplTriggers" form:"tplTriggers" binding:"omitempty,dive,required,max=255"`                // 分之推送自动触发合规 例如 ["commit"]

	KeyId models.Id `form:"keyId" json:"keyId" binding:"omitempty,startswith=k-,max=32"` // 部署密钥ID

}

type SearchTemplateForm struct {
	PageForm

	Q      string `form:"q" json:"q" binding:""`
	Status string `form:"status" json:"status" binding:"omitempty,oneof=enable disable"`
}

type UpdateTemplateForm struct {
	BaseForm
	Id           models.Id   `uri:"id" form:"id" json:"id" binding:"required,startswith=tpl-,max=32"`
	Name         string      `form:"name" json:"name" binding:"omitempty,gte=2,lte=64"`
	Description  string      `form:"description" json:"description" binding:"omitempty,max=255"`
	Status       string      `form:"status" json:"status" binding:"omitempty,oneof=enable disable"`
	Workdir      string      `form:"workdir" json:"workdir" binding:"max=255"`
	RunnerId     string      `json:"runnerId" form:"runnerId" binding:"max=255"`
	Playbook     string      `json:"playbook" form:"playbook" binding:"omitempty,endswith=.yml,max=255"`
	PlayVarsFile string      `json:"playVarsFile" form:"playVarsFile" binding:"max=255"`
	TfVarsFile   string      `form:"tfVarsFile" json:"tfVarsFile" binding:"max=255"`
	ProjectId    []models.Id `form:"projectId" json:"projectId" binding:"omitempty,dive,required,startswith=p-,max=32"`
	RepoRevision string      `form:"repoRevision" json:"repoRevision" binding:"max=64"`
	VcsId        models.Id   `form:"vcsId" json:"vcsId" binding:"omitempty,startswith=vcs-,max=32"`
	RepoId       string      `form:"repoId" json:"repoId" binding:"max=255"`
	RepoFullName string      `form:"repoFullName" json:"repoFullName" binding:"max=255"`
	TfVersion    string      `form:"tfVersion" json:"tfVersion" binding:"max=64"`

	Variables []Variable `json:"variables" form:"variables" binding:"omitempty,dive,required"`

	VarGroupIds    []models.Id `json:"varGroupIds" form:"varGroupIds" binding:"omitempty,dive,required,startswith=vg-,max=32"`
	DelVarGroupIds []models.Id `json:"delVarGroupIds" form:"delVarGroupIds" binding:"omitempty,dive,required,startswith=vg-,max=32"`
	PolicyEnable   bool        `json:"policyEnable" form:"policyEnable"`                                                        // 是否开启合规检测
	PolicyGroup    []models.Id `json:"policyGroup" form:"policyGroup" binding:"omitempty,dive,required,startswith=pog-,max=32"` // 绑定的合规策略组
	TplTriggers    []string    `json:"tplTriggers" form:"tplTriggers" binding:"omitempty,dive,required,max=255"`                // 分之推送自动触发合规 例如 ["commit"]
	KeyId          models.Id   `form:"keyId" json:"keyId" binding:"omitempty,startswith=k-,max=32"`                             // 部署密钥ID
}

type DeleteTemplateForm struct {
	BaseForm
	Id models.Id `uri:"id" json:"id" binding:"required,startswith=tpl-,max=32" swaggerignore:"true"`
}

type DetailTemplateForm struct {
	BaseForm
	Id models.Id `uri:"id" json:"id" binding:"required,startswith=tpl-,max=32" swaggerignore:"true"`
	// TODO 返回要返回 projectId
}

type OpenApiDetailTemplateForm struct {
	PageForm
	Id models.Id `form:"id" json:"id" binding:"required,startswith=tpl-,max=32"`
}

type OverviewTemplateForm struct {
	PageForm
	Id models.Id `form:"id" json:"id" binding:"required,startswith=tpl-,max=32"`
}

type RepoFileSearchForm struct {
	BaseForm
	RepoId       string    `json:"repoId" form:"repoId" binding:"required,max=255"`
	RepoRevision string    `json:"repoRevision" form:"repoRevision" binding:"required,max=64"`
	VcsId        models.Id `uri:"id" binding:"required,startswith=vcs-,max=32"`
	Workdir      string    `json:"workdir" form:"workdir" binding:"max=255"`
}

type TemplateVariableSearchForm struct {
	BaseForm
	RepoId       string    `json:"repoId" form:"repoId" binding:"required,max=255"`
	RepoRevision string    `json:"repoRevision" form:"repoRevision" binding:"required,max=64"`
	VcsId        models.Id `uri:"id" binding:"required,startswith=vcs-,max=32"`
	Workdir      string    `json:"workdir" form:"workdir" binding:"max=255"`
}

type TemplateTfVersionSearchForm struct {
	BaseForm
	VcsId     models.Id `json:"vcsId" form:"vcsId" binding:"required,startswith=vcs-,max=32"`
	VcsBranch string    `json:"vcsBranch" form:"vcsBranch" binding:"max=64"`
	RepoId    string    `json:"repoId" form:"repoId" binding:"max=255"`
}

type TemplateChecksForm struct {
	BaseForm
	Name         string    `json:"name" form:"name" binding:"omitempty,gte=2,lte=64"`
	RepoId       string    `json:"repoId" form:"repoId" binding:"max=255"`
	RepoRevision string    `json:"repoRevision" form:"repoRevision" binding:"max=64"`
	VcsId        models.Id `json:"vcsId" form:"vcsId" binding:"omitempty,startswith=vcs-,max=32"`
	Workdir      string    `json:"workdir" form:"workdir" binding:"max=255"`
	TemplateId   models.Id `json:"templateId" form:"templateId" binding:"omitempty,startswith=tpl-,max=32"`
	TfVarsFile   string    `json:"tfVarsFile" form:"tfVarsFile" binding:"max=255"`
	Playbook     string    `json:"playbook" form:"playbook" binding:"omitempty,endswith=.yml,max=255"`
}
