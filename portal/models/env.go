package models

import (
	"cloudiac/portal/libs/db"
	"database/sql/driver"
	"github.com/lib/pq"
	"path"
)

const (
	EnvStatusActive   = "active"   // 成功部署
	EnvStatusFailed   = "failed"   // apply 过程中出现错误
	EnvStatusInactive = "inactive" // 资源未部署或已销毁

	//EnvStatusDeploying = "deploying" // apply 运行中(plan 作业不改变状态)
	//EnvStatusApproving = "approving" // 等待审批
)

var (
	EnvStatus     = []string{EnvStatusActive, EnvStatusFailed, EnvStatusInactive}
	EnvTaskStatus = []string{TaskRunning, TaskApproving} // 环境 taskStatus 有效值
)

type Env struct {
	SoftDeleteModel
	OrgId     Id `json:"orgId" gorm:"size:32;not null"`     // 组织ID
	ProjectId Id `json:"projectId" gorm:"size:32;not null"` // 项目ID
	TplId     Id `json:"tplId" gorm:"size:32;not null"`     // 模板ID
	CreatorId Id `json:"creatorId" gorm:"size:32;not null"` // 创建人ID

	Name        string `json:"name" gorm:"not null"`                                                                       // 环境名称
	Description string `json:"description" gorm:"type:text"`                                                               // 环境描述
	Status      string `json:"status" gorm:"type:enum('active','failed','inactive')" enums:"'active','failed','inactive'"` // 环境状态, active活跃, inactive非活跃,failed错误,running部署中,approving审批中
	// 任务状态，只同步部署任务的状态(apply,destroy)，plan 任务不会对环境产生影响，所以不同步
	TaskStatus string `json:"taskStatus" gorm:"type:enum('', 'pending','approving','running');default:''"`
	Archived   bool   `json:"archived" gorm:"default:'0'"`                             // 是否已归档
	RunnerId   string `json:"runnerId" gorm:"size:32;not null"`                        //部署通道ID
	Timeout    int    `json:"timeout" gorm:"default:'600';comment:'部署超时'"`             // 部署超时时间（单位：秒）
	OneTime    bool   `json:"oneTime" gorm:"default:'0'"`                              // 一次性环境标识
	Deploying  bool   `json:"deploying" gorm:"not null;default:'0';common:'是否正在执行部署'"` // 是否正在执行部署

	StatePath string `json:"statePath" gorm:"not null" swaggerignore:"true"` // Terraform tfstate 文件路径（内部）

	// 环境可以覆盖模板中的 vars file 配置，具体说明见 Template model
	Variables    EnvVariables `json:"variables" gorm:"type:json"`               // 合并变量列表
	TfVarsFile   string       `json:"tfVarsFile" gorm:"default:''"`             // Terraform tfvars 变量文件路径
	PlayVarsFile string       `json:"playVarsFile" gorm:"default:''"`           // Ansible 变量文件路径
	Playbook     string       `json:"playbook" gorm:"default:''"`               // Ansible playbook 入口文件路径
	Revision     string       `json:"revision" gorm:"size:64;default:'master'"` // Vcs仓库分支/标签
	KeyId        Id           `json:"keyId" gorm:"size32"`                      // 部署密钥ID

	LastTaskId Id `json:"lastTaskId" gorm:"size:32"` // 最后一次部署或销毁任务的 id(plan 任务不记录)

	AutoApproval bool `json:"autoApproval" gorm:"default:'0'"` // 是否自动审批

	TTL           string `json:"ttl" gorm:"default:'0'" example:"1h/1d"` // 生命周期
	AutoDestroyAt *Time  `json:"autoDestroyAt"`                          // 自动销毁时间

	// 该 id 在创建自动销毁任务后保存，并在销毁任务执行完成后清除
	AutoDestroyTaskId Id `json:"-"  gorm:"default:''"` // 自动销毁任务 id

	// 触发器设置
	Triggers pq.StringArray `json:"triggers" gorm:"type:json" swaggertype:"array,string"` // 触发器。commit（每次推送自动部署），prmr（提交PR/MR的时候自动执行plan）
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
	return path.Join(e.OrgId.String(), e.ProjectId.String(), e.Id.String(), "terraform.tfstate")
}

func (e *Env) HideSensitiveVariable() {
	for index, v := range e.Variables {
		if v.Sensitive {
			e.Variables[index].Value = ""
		}
	}
}

func (e *Env) MergeTaskStatus() string {
	if e.Deploying {
		e.Status = e.TaskStatus
	}
	return e.Status
}

type EnvVariables []Variable

func (v EnvVariables) Value() (driver.Value, error) {
	return MarshalValue(v)
}

func (v *EnvVariables) Scan(value interface{}) error {
	return UnmarshalValue(value, v)
}

type EnvDetail struct {
	Env
	Creator       string `json:"creator"`       // 创建人
	ResourceCount int    `json:"resourceCount"` // 资源数量
	TemplateName  string `json:"templateName"`  // 模板名称
	KeyName       string `json:"keyName"`       // 密钥名称
}
