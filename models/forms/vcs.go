package forms

type CreateVcsForm struct {
	BaseForm
	Name     string `form:"name" json:"name" binding:"required"`
	VcsType  string `form:"vcs_type" json:"vcs_type" binding:"required"`
	Address  string `form:"address" json:"address" binding:"required"`
	VcsToken string `form:"vcs_token" json:"vcs_token" binding:"required"`
	Status   string `form:"status" json:"status" binding:"required"`
}

type UpdateVcsForm struct {
	BaseForm
	Id     uint   `form:"id" json:"id" binding:"required"`
	Status string `form:"status" json:"status" binding:"required"`
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
