package forms

type Var struct {
	Key      string `form:"key" json:"key" binding:"required"`
	Value    string `form:"value" json:"value" binding:"required"`
	IsSecret bool   `form:"isSecret" json:"isSecret" binding:"required,default:false"`
}

type CreateTemplateForm struct {
	BaseForm
	Name        string `form:"name" json:"name" binding:"required,gte=2,lte=32"`
	OrgId       int    `form:"orgId" json:"orgId" binding:"required"`
	Description string `form:"description" json:"Description" binding:"required"`
	RepoId      int    `form:"repoId" json:"repoId" binding:"required"`
	RepoAddr    string `form:"repoAddr" json:"repoAddr" bingding:"required"`
	RepoBranch  string `form:"repoBranch" json:"repoBranch" bingding:"required"`
	SaveState   bool   `form:"saveState" json:"saveState" binding:"required,default:false`
	Vars        []Var  `form:"vars" json:"vars"`
	Varfile     string `form:"varfile" json:"varfile"`
	Extra       string `form:"extra" json:"extra"`
	Timeout     int    `form:"timeout" json:"timeout"`
}

type SearchTemplateForm struct {
	BaseForm

	Q      string `form:"q" json:"q" binding:""`
	Status string `form:"status" json:"status"`
}

type UpdateTemplateForm struct {
	BaseForm
	Id          uint   `form:"id" json:"id" binding:""`
	Name        string `form:"name" json:"name" binding:"gte=2,lte=32"`
	Description string `form:"description" json:"Description" binding:"required"`
	Vars        []Var  `form:"vars" json:"vars"`
	Varfile     string `form:"varfile" json:"varfile"`
	Extra       string `form:"extra" json:"extra"`
	Timeout     int    `form:"timeout" json:"timeout"`
}
