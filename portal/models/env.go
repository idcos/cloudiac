package models

import (
	"cloudiac/portal/libs/db"
	"fmt"
	"path"
	"time"
)

const (
	EnvStatusActive   = "active"   // 成功部署
	EnvStatusFailed   = "failed"   // apply 过程中出现错误
	EnvStatusInactive = "inactive" // 资源未部署或已销毁

	//EnvStatusDeploying = "deploying" // apply 运行中(plan 作业不改变状态)
	//EnvStatusApproving = "approving" // 等待审批
)

var EnvStatus = []string{EnvStatusActive, EnvStatusFailed, EnvStatusInactive}

type Env struct {
	SoftDeleteModel
	OrgId     Id `json:"orgId" gorm:"size:32;not null"`     // 组织ID
	ProjectId Id `json:"projectId" gorm:"size:32;not null"` // 项目ID
	TplId     Id `json:"tplId" gorm:"size:32;not null"`     // 模板ID
	CreatorId Id `json:"creatorId" gorm:"size:32;not null"` // 创建人ID

	Name        string `json:"name" gorm:"not null"`                                                                       // 环境名称
	Description string `json:"description" gorm:"type:text"`                                                               // 环境描述
	Status      string `json:"status" gorm:"type:enum('active','failed','inactive')" enums:"'active','failed','inactive'"` // 环境状态
	// 任务状态，只同步部署任务的状态(apply,destroy)，plan 任务不会对环境产生影响，所以不同步
	TaskStatus string `json:"taskStatus" gorm:"type:enum('', 'pending','approving','running');default:''"`
	Archived   bool   `json:"archived" gorm:"default:'0'"`                             // 是否已归档
	RunnerId   string `json:"runnerId" gorm:"size:32;not null"`                        //部署通道ID
	Timeout    int    `json:"timeout" gorm:"default:'600';comment:'部署超时'"`             // 部署超时时间（单位：秒）
	OneTime    bool   `json:"oneTime" gorm:"default:'0'"`                              // 一次性环境标识
	Deploying  bool   `json:"deploying" gorm:"not null;default:'0';common:'是否正在执行部署'"` // 是否正在执行部署

	StatePath string `json:"statePath" gorm:"not null" swaggerignore:"true"` // Terraform tfstate 文件路径（内部）
	Outputs   string `json:"outputs" gorm:"type:text" swaggerignore:"true"`  // Terraform outputs 输出内容

	// 环境可以覆盖模板中的 vars file 配置，具体说明见 Template model
	Variables    []VariableBody `json:"variables" gorm:"json"`          // 合并变量列表
	TfVarsFile   string         `json:"tfVarsFile" gorm:"default:''"`   // Terraform tfvars 变量文件路径
	PlayVarsFile string         `json:"playVarsFile" gorm:"default:''"` // Ansible 变量文件路径
	Playbook     string         `json:"playbook" gorm:"default:''"`     // Ansible playbook 入口文件路径

	LastTaskId Id `json:"lastTaskId" gorm:"size:32"` // 最后一次部署或销毁任务的 id(plan 任务不记录)

	// AutoDestroyAt 自动销毁时间，这里存绝对时间，1小时、2小时的相对时间选择由前端转换
	// TODO 自动销毁机制待实现
	TTL           int        `json:"ttl" gorm:"default:'0'"`          // 生存时间
	AutoDestroyAt *time.Time `json:"autoDestroyAt"`                   // 自动销毁时间
	AutoApproval  bool       `json:"autoApproval" gorm:"default:'0'"` // 是否自动审批
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

func (e *Env) DefaultStatPath() string {
	return path.Join(fmt.Sprintf("env-%s", e.Id.String()), "terraform.tfstate")
}

// EnvRes 环境资源
// 环境资源为该环境部署后 terraform 创建的资源列表
type EnvRes struct {
	BaseModel

	OrgId     Id `json:"orgId" gorm:"size:32;not null"`     // 组织ID
	ProjectId Id `json:"projectId" gorm:"size:32;not null"` // 项目ID
	EnvId     Id `json:"envId" gorm:"size:32;not null"`     // 环境ID

	Provider string `json:"provider" gorm:"not null"` // Terraform provider，一般表示为云商/云平台
	Type     string `json:"type" gorm:"not null"`     // 资源类型
	Name     string `json:"name" gorm:"not null"`     // 资源名称
	Index    int    `json:"index" gorm:"not null"`    // 资源序号
	Attrs    JSON   `json:"attrs" gorm:"type:text"`   // 资源属性
}

func (EnvRes) TableName() string {
	return "iac_env_res"
}

type EnvDetail struct {
	Env
	Creator       string `json:"creator"`       // 创建人
	ResourceCount int    `json:"resourceCount"` // 资源数量
	TemplateName  string `json:"templateName"`  // 模板名称
}
