package models

import (
	"cloudiac/portal/libs/db"
	"time"
)

const (
	EnvStatusActive   = "active"   // 成功部署
	EnvStatusFailed   = "failed"   // apply 过程中出现错误
	EnvStatusInactive = "inactive" // 资源未部署或已销毁

	//EnvStatusDeploying = "deploying" // apply 运行中(plan 作业不改变状态)
	//EnvStatusApproving = "approving" // 等待审批
)

type Env struct {
	SoftDeleteModel
	OrgId     Id `json:"orgId" gorm:"size:32;not null"`
	TplId     Id `json:"tplId" gorm:"size:32;not null"`
	ProjectId Id `json:"projectId" gorm:"size:32;not null"`

	Name        string `json:"name" gorm:"not null"`
	Description string `json:"description" gorm:"type:text"`

	Status string `json:"status" gorm:"type:enum('active','failed','inactive')"`
	// 任务状态，只同步部署任务的状态(apply,destroy)，plan 任务不会对环境产生影响，所以不同步
	TaskStatus string `json:"taskStatus" gorm:"type:enum('', 'pending','approving','running');default:''"`
	Deploying  bool   `json:"Deploying" gorm:"not null;default:'0';common:'是否正在执行部署'"`

	RunnerId string `json:"runnerId" gorm:"size:32;not null"`
	Timeout  int    `json:"timeout" gorm:"default:'600';comment:'部署超时'"`
	OneTime  bool   `json:"oneTime" gorm:"default:'0'"`

	StatePath string `json:"statePath" gorm:"not null"`

	// 环境可以覆盖模板中的 vars file 配置，具体说明见 Template model
	TfVarsFile   string `json:"tfVarsFile" gorm:"default:''"`
	PlayVarsFile string `json:"playVarsFile" gorm:"default:''"`

	// 最后一次部署或销毁任务的 id(plan 作业不记录)
	LastTaskId Id `json:"lastTaskId" gorm:"size:32;default:''"`

	// TODO 自动销毁机制待实现
	TTL           int        `json:"ttl" gorm:"default:'0'"` // 生存时间
	AutoDestroyAt *time.Time `json:"autoDestroyAt"`          // 自动销毁时间

	AutoApprove bool `json:"autoApprove" gorm:"default:'0'"`
}

func (Env) TableName() string {
	return "iac_env"
}

func (e *Env) Migrate(sess *db.Session) (err error) {
	if err = sess.RemoveIndex("iac_env", "unique__tpl__env__name"); err != nil {
		return err
	}
	if err = e.AddUniqueIndex(sess, "unique__project__env__name",
		"project_id", "name"); err != nil {
		return err
	}
	return nil
}
