package forms

type GetGitProjectsForm struct {
	BaseForm
	Q            string    `form:"q" json:"q"`
	Url          string    `form:"url" json:"url"`
	Type 		 string    `json:"type"`
	Token        string    `json:"token"`
}

type GetGitBranchesForm struct {
	BaseForm
	RepoId       int    `form:"repoId" json:"repoId"`
	Url          string    `form:"url" json:"url"`
	Token        string    `json:"token"`
	Type 		 string    `json:"type"`
}

type GetReadmeForm struct {
	BaseForm
	RepoId       int    `form:"repoId" json:"repoId" binding:"required"`
	Branch       string `form:"branch" json:"branch"`
	Url          string    `form:"url" json:"url"`
	Token        string    `json:"token"`
	Type 		 string    `json:"type"`
}