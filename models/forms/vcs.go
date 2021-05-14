package forms

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
	Id     uint   `form:"id" json:"id" binding:"required"`
	Status string `form:"status" json:"status" binding:""`
	Name     string `form:"name" json:"name" binding:""`
	VcsType  string `form:"vcsType" json:"vcsType" binding:""`
	Address  string `form:"address" json:"address" binding:""`
	VcsToken string `form:"vcsToken" json:"vcsToken" binding:""`
}

type SearchVcsForm struct {
	BaseForm
	Q      string `form:"q" json:"q" binding:""`
	Status string `form:"status" json:"status"`
}

type DeleteVcsForm struct {
	BaseForm
	Id uint `form:"id" json:"id" binding:"required"`
}
