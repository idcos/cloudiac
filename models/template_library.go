package models

type TemplateLibrary struct {
	SoftDeleteModel
	Name                   string `json:"name" gorm:"size:32;not null;comment:'模版名称'"`
	Description            string `json:"description" gorm:"size:255;comment:'描述'"`
	RepoId                 int    `json:"repoId" gorm:"size:32;comment:'仓库ID'"`
	RepoAddr               string `json:"repoAddr" gorm:"size:128;default:'';comment:'仓库地址'"`
	RepoBranch             string `json:"repoBranch" gorm:"size:64;default:'master';comment:'仓库分支'"`
	SaveState              bool   `json:"saveState" gorm:"defalut:false;comment:'是否保存状态'"`
	Vars                   JSON   `json:"vars" gorm:"type:json;null;comment:'变量'"`
	Varfile                string `json:"varfile" gorm:"size:128;default:'';comment:'变量文件'"`
	Timeout                int64  `json:"timeout" gorm:"default:300;comment:'超时时长'"`
}

func (TemplateLibrary) TableName() string {
	return "iac_template_library"
}

