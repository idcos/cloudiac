// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package forms

import "cloudiac/portal/models"

type CreateVcsForm struct {
	BaseForm
	Name     string `form:"name" json:"name" binding:"required,lte=255"`
	VcsType  string `form:"vcsType" json:"vcsType" binding:"required,lte=255"`
	Address  string `form:"address" json:"address" binding:"required,lte=255"`
	VcsToken string `form:"vcsToken" json:"vcsToken" binding:"required,lte=255"`
}

type UpdateVcsForm struct {
	BaseForm
	Id       models.Id `uri:"id" json:"id" binding:"" swaggerignore:"true"`
	Status   string    `form:"status" json:"status" binding:""`
	Name     string    `form:"name" json:"name" binding:""`
	VcsType  string    `form:"vcsType" json:"vcsType" binding:""`
	Address  string    `form:"address" json:"address" binding:""`
	VcsToken string    `form:"vcsToken" json:"vcsToken" binding:""`
}

type SearchVcsForm struct {
	NoPageSizeForm

	Q                string `form:"q" json:"q" binding:""`
	Status           string `form:"status" json:"status"`
	IsShowDefaultVcs bool   `form:"isShowDefaultVcs" json:"isShowDefaultVcs" default:"true"`
}

type DeleteVcsForm struct {
	BaseForm
	Id models.Id `uri:"id" json:"id" binding:"required" swaggerignore:"true"`
}

type GetGitProjectsForm struct {
	PageForm
	Id models.Id `uri:"id" json:"id" binding:"" swaggerignore:"true"`
	Q  string    `form:"q" json:"q"`
}

type GetGitRevisionForm struct {
	BaseForm
	Id     models.Id `uri:"id" json:"id" binding:"" swaggerignore:"true"`
	RepoId string    `form:"repoId" json:"repoId" binding:"required"`
}

type GetReadmeForm struct {
	BaseForm
	Id           models.Id `uri:"id" json:"id" binding:"" swaggerignore:"true"`
	RepoId       string    `form:"repoId" json:"repoId" binding:"required"`
	RepoRevision string    `json:"repoRevision" form:"repoRevision" binding:"required"`
	Dir          string    `form:"dir" json:"dir"` // 指定目录名，默认读取根目录
}

type GetVcsRepoFileForm struct {
	BaseForm
	Id       models.Id `uri:"id" json:"id" binding:"" swaggerignore:"true"`
	RepoId   string    `form:"repoId" json:"repoId" binding:"required"`
	Branch   string    `form:"branch" json:"branch" binding:"required"`
	FileName string    `json:"fileName" form:"fileName" binding:"required"`
	Workdir  string    `json:"workdir" form:"workdir"`
}
