package forms

type GetGitProjectsForm struct {
	BaseForm
	Q            string    `form:"q" json:"q"`
	VcsId 	     uint `form:"vcsId" json:"vcsId" binding:"required"`

}

type GetGitBranchesForm struct {
	BaseForm
	RepoId       int    `form:"repoId" json:"repoId"`
	VcsId 	     uint `form:"vcsId" json:"vcsId" binding:"required"`

}

type GetReadmeForm struct {
	BaseForm
	RepoId       int    `form:"repoId" json:"repoId" binding:"required"`
	Branch       string `form:"branch" json:"branch"`
	VcsId 	     uint `form:"vcsId" json:"vcsId" binding:"required"`

}