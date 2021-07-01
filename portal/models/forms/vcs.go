package forms

import "cloudiac/portal/models"

type CreateVcsForm struct {
	BaseForm
	Name     string `form:"name" json:"name" binding:"required"`
	VcsType  string `form:"vcsType" json:"vcsType" binding:"required"`
	Address  string `form:"address" json:"address" binding:"required"`
	VcsToken string `form:"vcsToken" json:"vcsToken" binding:"required"`
	Status   string `form:"status" json:"status" binding:"required"`
}

type UpdateVcsForm struct {
	BaseForm
	Id       models.Id `form:"id" json:"id" binding:"required"`
	Status   string    `form:"status" json:"status" binding:""`
	Name     string    `form:"name" json:"name" binding:""`
	VcsType  string    `form:"vcsType" json:"vcsType" binding:""`
	Address  string    `form:"address" json:"address" binding:""`
	VcsToken string    `form:"vcsToken" json:"vcsToken" binding:""`
}

type SearchVcsForm struct {
	BaseForm
	Q      string `form:"q" json:"q" binding:""`
	Status string `form:"status" json:"status"`
}

type DeleteVcsForm struct {
	BaseForm
	Id models.Id `form:"id" json:"id" binding:"required"`
}

type GetGitProjectsForm struct {
	BaseForm
	Q     string    `form:"q" json:"q"`
	VcsId models.Id `form:"vcsId" json:"vcsId" binding:"required"`
}

type GetGitBranchesForm struct {
	BaseForm
	RepoId string    `form:"repoId" json:"repoId"`
	VcsId  models.Id `form:"vcsId" json:"vcsId" binding:"required"`
}

type GetReadmeForm struct {
	BaseForm
	RepoId string    `form:"repoId" json:"repoId" binding:""`
	Branch string    `form:"branch" json:"branch" binding:"required"`
	VcsId  models.Id `form:"vcsId" json:"vcsId" binding:"required"`
}
