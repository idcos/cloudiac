package forms

type Var struct {
	Id          string `form:"id" json:"id" binding:"required"`
	Key         string `form:"key" json:"key" binding:"required"`
	Value       string `form:"value" json:"value" binding:"required"`
	IsSecret    *bool  `form:"isSecret" json:"isSecret" binding:"required,default:false"`
	Type        string `form:"type" json:"type" binding:"required,default:env"`
	Description string `form:"description" json:"description" binding:""`
}

type CreateTemplateForm struct {
	BaseForm
	Name        string `form:"name" json:"name" binding:"required,gte=2,lte=32"`
	Description string `form:"description" json:"Description" binding:""`
	RepoId      int    `form:"repoId" json:"repoId" binding:"required"`
	RepoAddr    string `form:"repoAddr" json:"repoAddr" bingding:"required"`
	RepoBranch  string `form:"repoBranch" json:"repoBranch" bingding:"required"`
	SaveState   *bool  `form:"saveState" json:"saveState"`
	Vars        []Var  `form:"vars" json:"vars"`
	Varfile     string `form:"varfile" json:"varfile"`
	Extra       string `form:"extra" json:"extra"`
	Timeout     int64  `form:"timeout" json:"timeout"`
}

type SearchTemplateForm struct {
	BaseForm

	Q          string `form:"q" json:"q" binding:""`
	Status     string `form:"status" json:"status"`
	TaskStatus string `json:"taskStatus" form:"taskStatus" `
}

type UpdateTemplateForm struct {
	BaseForm
	Id          uint   `form:"id" json:"id" binding:"required"`
	Name        string `form:"name" json:"name"`
	Description string `form:"description" json:"Description"`
	SaveState   bool   `form:"saveState" json:"saveState"`
	Vars        []Var  `form:"vars" json:"vars"`
	Varfile     string `form:"varfile" json:"varfile"`
	Extra       string `form:"extra" json:"extra"`
	Timeout     int    `form:"timeout" json:"timeout"`
	Status      string `form:"status" json:"status"`
}

type DetailTemplateForm struct {
	BaseForm
	Id uint `form:"id" json:"id" binding:"required"`
}

type OverviewTemplateForm struct {
	BaseForm
	Id uint `form:"id" json:"id" binding:"required"`
}
