package forms

// SearchEnvDriftsForm 漂移结果列表查询
type SearchEnvDriftsForm struct {
	NoPageSizeForm
	Keyword string `json:"keyword" form:"keyword" binding:"max=255"`
	IsDrift *bool  `json:"isDrift" form:"isDrift"`
}
