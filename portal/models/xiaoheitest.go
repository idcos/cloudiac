package models

import (
	"database/sql/driver"
	"github.com/lib/pq"
	"time"
)

type TfModule struct {
	BaseModel
	Namespace string
	Provider  string
	Name      string
	Icon      string // 图标路径(参考图标库实现方案)

	VcsId  Id
	RepoId string // 仓库在 vcs 下的唯一标识

	CreatorId Id
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

func (TfModule) TableName() string {
	return "iac_tf_module"
}

// namespace + provider + name 添加唯一索引

type TfModuleVersion struct {
	BaseModel

	Namespace string // 添加 namespace 字段以方便做数据隔离
	ModuleId  Id     // tf_module 表的 id
	GitTag    string //  eg. 'v0.1.1' or '0.1.1'
	CommitId  string // 发布时 tag 对应的 commit id
	Version   string //  eg. '0.1.1'

	// inputs，outputs 为解析后的结构化数据，展示效果由前端渲染
	Inputs  TfModuleVariables ` gorm:"type:json"`
	Outputs TfModuleVariables ` gorm:"type:json"`

	ModuleDeps   TfModuleDependencies ` gorm:"type:json"`
	ProviderDeps TfModuleDependencies ` gorm:"type:json"`

	Resources pq.StringArray ` gorm:"type:json"` // 资源名称列表，展示效果由前端渲染

	CreatorId Id
	CreatedAt *time.Time
}

func (TfModuleVersion) TableName() string {
	return "iac_tf_module_version"
}

// moduleId + version 添加唯一索引

type TfSubModule struct {
	BaseModel
	VersionId Id // TfModuleVersion id
	Name      string

	Inputs  TfModuleVariables ` gorm:"type:json"`
	Outputs TfModuleVariables ` gorm:"type:json"`

	ModuleDeps   TfModuleDependencies ` gorm:"type:json"`
	ProviderDeps TfModuleDependencies ` gorm:"type:json"`

	Resources pq.StringArray ` gorm:"type:json"`
}

func (TfSubModule) TableName() string {
	return "iac_tf_sub_module"
}

type TfModuleExample struct {
	BaseModel
	VersionId Id
	Name      string

	Inputs  TfModuleVariables ` gorm:"type:json"`
	Outputs TfModuleVariables ` gorm:"type:json"`
}

func (TfModuleExample) TableName() string {
	return "iac_tf_module_example"
}

// model 内嵌套的 struct

type TfModuleVariables []TfModuleVariable

func (v TfModuleVariables) Value() (driver.Value, error) {
	return MarshalValue(v)
}

func (v *TfModuleVariables) Scan(value interface{}) error {
	return UnmarshalValue(value, v)
}

type TfModuleVariable struct {
	Name        string
	Type        string
	Default     interface{}
	Description string
	Sensitive bool        `json:"sensitive,omitempty"`
}

type TfModuleDependencies []TfModuleDependency

func (v TfModuleDependencies) Value() (driver.Value, error) {
	return MarshalValue(v)
}

func (v *TfModuleDependencies) Scan(value interface{}) error {
	return UnmarshalValue(value, v)
}

// moduel 的 module 或 provider 依赖
type TfModuleDependency struct {
	Name    string
	Source  string
	Version string // 版本号或者版本约束(eg. "1.0.0", ">=1.0.0")
}



// 计数表
type NumberCount struct {
	BaseModel
	Type     string // module_download/policy_download/provider_download
	ObjectId Id     // 计数对象的id(module/policy id)
	Count    int
}

func (NumberCount) TableName() string {
	return "iac_number_count"
}
