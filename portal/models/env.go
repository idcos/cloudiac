package models

import (
	"cloudiac/portal/libs/db"
	"time"
)

const (
	EnvStatusActive    = "active"    // 成功部署
	EnvStatusDeploying = "deploying" // apply 运行中(plan 作业不改变状态)
	EnvStatusApproving = "approving" // 等待审批
	EnvStatusFailed    = "failed"    // apply 过程中出现错误
	EnvStatusInactive  = "inactive"  // 资源未部署或已销毁
)

type Env struct {
	SoftDeleteModel
	OrgId     Id `json:"orgId" gorm:"size:32;not null"`
	ProjectId Id `json:"projectId" gorm:"size:32;not null"`
	TplId     Id `json:"tplId" gorm:"size:32;not null"`

	Name        string `json:"name" gorm:"not null"`
	Description string `json:"description" gorm:"type:text"`
	Status      string `json:"status" gorm:"type:enum('active','deploying','approving','failed','inactive')"`
	RunnerId    string `json:"runnerId" gorm:"size:32;not null"`
	Timeout     int    `json:"timeout" gorm:"default:'600';comment:'部署超时'"`
	OneTime     bool   `json:"oneTime" gorm:"default:'0'"`

	StatePath string `json:"statePath" gorm:"not null"`
	Outputs   string `json:"outputs" gorm:"type:text"`

	// 环境可以覆盖模板中的 vars file 配置，具体说明见 Template model
	TfVarsFile   string `json:"tfVarsFile" gorm:"default:''"`
	PlayVarsFile string `json:"playVarsFile" gorm:"default:''"`

	// 最后一次部署或销毁任务的 id(plan 作业不记录)
	LastTaskId Id `json:"lastTaskId" gorm:"size:32;default:'0'"`

	// AutoDestroyAt 自动销毁时间，这里存绝对时间，1小时、2小时的相对时间选择由前端转换
	AutoDestroyAt *time.Time `json:"autoDestroyAt"`
	AutoApproval  bool       `json:"autoApproval" gorm:"default:'0'"`
}

func (Env) TableName() string {
	return "iac_env"
}

func (e *Env) Migrate(sess *db.Session) (err error) {
	if err := e.AddUniqueIndex(sess, "unique__tpl__env__name", "tpl_id", "name"); err != nil {
		return err
	}
	return nil
}

type EnvRes struct {
	BaseModel

	OrgId     Id `json:"orgId" gorm:"size:32;not null"`
	ProjectId Id `json:"projectId" gorm:"size:32;not null"`
	EnvId     Id `json:"envId" gorm:"size:32;not null"`

	Provider string `json:"provider" gorm:"not null"`
	Type     string `json:"type" gorm:"not null"`
	Name     string `json:"name" gorm:"not null"`
	Index    int    `json:"index" gorm:"not null"`
	Attrs    JSON   `json:"attrs" gorm:"type:text"`
}

func (EnvRes) TableName() string {
	return "iac_env_res"
}
