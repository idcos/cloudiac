package forms

type GetGitProjectsForm struct {
	BaseForm
	Q            string    `form:"q" json:"q"`
}

type GetGitBranchesForm struct {
	BaseForm
	RepoId       int    `form:"repoId" json:"repoId"`
}

type GetReadmeForm struct {
	BaseForm
	RepoId       int    `form:"repoId" json:"repoId" binding:"required"`
	Branch       string `form:"branch" json:"branch"`
}