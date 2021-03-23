package forms

type CreateOrganizationForm struct {
	BaseForm
	Name           string `form:"name" json:"name" binding:"required,gte=2,lte=32"`
	Description    string `form:"description" json:"description" binding:""`
	VcsType        string `form:"vcsType" json:"vcsType" binding:""`
	VcsVersion     string `form:"vcsVersion" json:"vcsVersion" binding:""`
	VcsAuthInfo    string `form:"vcsAuthInfo" json:"vcsAuthInfo" binding:""`
}

type UpdateOrganizationForm struct {
	BaseForm
	Id             uint   `form:"id" json:"id" binding:""`
	Name           string `form:"name" json:"name" binding:""`
	Description    string `form:"description" json:"description" binding:"max=255"`
	VcsType        string `form:"vcsType" json:"vcsType" binding:""`
	VcsVersion     string `form:"vcsVersion" json:"vcsVersion" binding:""`
	VcsAuthInfo    string `form:"vcsAuthInfo" json:"vcsAuthInfo" binding:""`
}

type SearchOrganizationForm struct {
	BaseForm

	Q          string `form:"q" json:"q" binding:""`
	Status     string    `form:"status" json:"status"`
}

type DeleteOrganizationForm struct {
	BaseForm
	Id uint `form:"id" json:"id" binding:"required"`
}

type DisableOrganizationForm struct {
	BaseForm

	Id     uint `form:"id" json:"id" binding:"required"`
	Status string  `form:"status" json:"status" binding:"required"`
}

type DetailOrganizationForm struct {
	BaseForm
	Id uint `form:"id" json:"id" binding:"required"`
}
