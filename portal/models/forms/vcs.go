// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package forms

import "cloudiac/portal/models"

type CreateVcsForm struct {
	BaseForm
	Name     string `form:"name" json:"name" binding:"required,gte=2,lte=255"`
	VcsType  string `form:"vcsType" json:"vcsType" binding:"required,lte=255"`
	Address  string `form:"address" json:"address" binding:"required,lte=255"`
	VcsToken string `form:"vcsToken" json:"vcsToken" binding:"required,lte=255"`
}

type UpdateVcsForm struct {
	BaseForm
	Id       models.Id `uri:"id" json:"id" binding:"required,max=32" swaggerignore:"true"`
	Status   string    `form:"status" json:"status" binding:"omitempty,oneof=enable disable"`
	Name     string    `form:"name" json:"name" binding:"omitempty,lte=255"`
	VcsType  string    `form:"vcsType" json:"vcsType" binding:"omitempty,lte=255"`
	Address  string    `form:"address" json:"address" binding:"omitempty,lte=255"`
	VcsToken string    `form:"vcsToken" json:"vcsToken" binding:"omitempty,lte=255"`
}

type SearchVcsForm struct {
	NoPageSizeForm

	Q                string `form:"q" json:"q" binding:""`
	Status           string `form:"status" json:"status" binding:"omitempty,oneof=enable disable"`
	IsShowDefaultVcs bool   `form:"isShowDefaultVcs" json:"isShowDefaultVcs" default:"true"`
}

type DeleteVcsForm struct {
	BaseForm
	Id models.Id `uri:"id" json:"id" binding:"required,max=32" swaggerignore:"true"`
}

type GetGitProjectsForm struct {
	PageForm
	Id models.Id `uri:"id" json:"id" binding:"required,max=32" swaggerignore:"true"`
	Q  string    `form:"q" json:"q"`
}

type GetGitRevisionForm struct {
	BaseForm
	Id     models.Id `uri:"id" json:"id" binding:"required,max=32" swaggerignore:"true"`
	RepoId string    `form:"repoId" json:"repoId" binding:"required"`
}

type GetReadmeForm struct {
	BaseForm
	Id           models.Id `uri:"id" json:"id" binding:"required,max=32" swaggerignore:"true"`
	RepoId       string    `form:"repoId" json:"repoId" binding:"required,max=255"`
	RepoRevision string    `json:"repoRevision" form:"repoRevision" binding:"required,max=64"`
	Dir          string    `form:"dir" json:"dir" binding:"max=255"` // 指定目录名，默认读取根目录
}

type GetVcsRepoFileForm struct {
	BaseForm
	Id       models.Id `uri:"id" json:"id" binding:"required,max=32" swaggerignore:"true"`
	RepoId   string    `form:"repoId" json:"repoId" binding:"required,max=255"`
	Branch   string    `form:"branch" json:"branch" binding:"required,max=255"`
	FileName string    `json:"fileName" form:"fileName" binding:"required,max=255"`
	Workdir  string    `json:"workdir" form:"workdir" binding:"max=255"`
}

type GetFileFullPathForm struct {
	BaseForm
	Id           models.Id `uri:"id" json:"id" binding:"required,max=32" swaggerignore:"true"`
	RepoId       string    `form:"repoId" json:"repoId" binding:"required,max=255"`
	RepoRevision string    `json:"repoRevision" form:"repoRevision" binding:"required,max=64"`
	Path         string    `json:"path" form:"path"` // 文件路径 如workdir/test.tfvars
	CommitId     string    `json:"commitId" form:"commitId"`
}
