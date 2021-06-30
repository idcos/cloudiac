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
	Guid      string `json:"guid" gorm:"size:32;not null;unique"`
	OrgId     uint   `json:"orgId" gorm:"not null"`
	ProjectId uint   `json:"projectId" gorm:"not null"`
	TplId     uint   `json:"tplId" gorm:"not null"`

	Name        string `json:"name" gorm:"not null"`
	Description string `json:"description" gorm:"type:text"`
	Status      string `json:"status" gorm:"type:enum('active','deploying','approving','failed','inactive')"`
	RunnerId    string `json:"runnerId"`
	Timeout     int64  `json:"timeout" gorm:"default:'300';comment:'超时时长'"`
	OneTime     bool   `json:"oneTime" gorm:"default:'0'"`

	// 最后一次部署或销毁任务的 id(plan 作业不记录)
	LastTaskId uint `json:"lastTaskId" gorm:"default:'0'"`

	// AutoDestroyAt 自动销毁时间，这里存绝对时间，1小时、2小时的相对时间选择由前端转换
	AutoDestroyAt *time.Time `json:"autoDestroyAt"`
	AutoApproval  bool       `json:"autoApproval" gorm:"default:'0'"`

	StatePath string `json:"statePath" gorm:"not null"`
	Outputs   string `json:"outputs" gorm:"type:text"`
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

	Guid      string `json:"guid" gorm:"size:32;not null;unique"`
	OrgId     uint   `json:"orgId" gorm:"not null"`
	ProjectId uint   `json:"projectId" gorm:"not null"`
	EnvId     uint   `json:"envId" gorm:"not null"`

	Provider string `json:"provider" gorm:"not null"`
	Type     string `json:"type" gorm:"not null"`
	Name     string `json:"name" gorm:"not null"`
	Index    uint   `json:"index" gorm:"not null"`
	Attrs    JSON   `json:"attrs" gorm:"type:text"`
}

func (EnvRes) TableName() string {
	return "iac_env_res"
}
