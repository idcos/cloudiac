// Copyright 2021 CloudJ Company Limited. All rights reserved.

package models

import (
	"cloudiac/portal/libs/db"
	"path"
	"time"

	"github.com/lib/pq"
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
	TaskStatus string `json:"taskStatus" gorm:"type:enum('','approving','running');default:''"`
	Archived   bool   `json:"archived" gorm:"default:false"`           // 是否已归档
	Timeout    int    `json:"timeout" gorm:"default:600;comment:部署超时"` // 部署超时时间（单位：秒）
	OneTime    bool   `json:"oneTime" gorm:"default:false"`            // 一次性环境标识
	Deploying  bool   `json:"deploying" gorm:"not null;default:false"` // 是否正在执行部署

	StatePath string `json:"statePath" gorm:"not null" swaggerignore:"true"` // Terraform tfstate 文件路径（内部）

	// 环境可以覆盖模板中的 vars file 配置，具体说明见 Template model
	TfVarsFile   string `json:"tfVarsFile" gorm:"default:''"`   // Terraform tfvars 变量文件路径
	PlayVarsFile string `json:"playVarsFile" gorm:"default:''"` // Ansible 变量文件路径
	Playbook     string `json:"playbook" gorm:"default:''"`     // Ansible playbook 入口文件路径

	// 任务相关参数，获取详情的时候，如果有 last_task_id 则返回 last_task_id 相关参数
	RunnerId string `json:"runnerId" gorm:"size:32;not null"`         //部署通道ID
	Revision string `json:"revision" gorm:"size:64;default:'master'"` // Vcs仓库分支/标签
	KeyId    Id     `json:"keyId" gorm:"size32"`                      // 部署密钥ID

	LastTaskId    Id `json:"lastTaskId" gorm:"size:32"`    // 最后一次部署或销毁任务的 id(plan 任务不记录)
	LastResTaskId Id `json:"lastResTaskId" gorm:"size:32"` // 最后一次进行了资源列表统计的部署任务的 id

	LastScanTaskId Id `json:"lastScanTaskId" gorm:"size:32"` // 最后一次策略扫描任务 id

	AutoApproval    bool `json:"autoApproval" gorm:"default:false"`    // 是否自动审批
	StopOnViolation bool `json:"stopOnViolation" gorm:"default:false"` // 当合规不通过是否中止部署

	TTL           string `json:"ttl" gorm:"default:'0'" example:"1h/1d"` // 生命周期
	AutoDestroyAt *Time  `json:"autoDestroyAt" gorm:"type:datetime"`     // 自动销毁时间

	// 该 id 在创建自动销毁任务后保存，并在销毁任务执行完成后清除
	AutoDestroyTaskId Id `json:"-"  gorm:"default:''"` // 自动销毁任务 id

	// 触发器设置
	Triggers pq.StringArray `json:"triggers" gorm:"type:text" swaggertype:"array,string"` // 触发器。commit（每次推送自动部署），prmr（提交PR/MR的时候自动执行plan）

	// 任务重试
	RetryNumber int  `json:"retryNumber" gorm:"size:32;default:3"` // 任务重试次数
	RetryDelay  int  `json:"retryDelay" gorm:"size:32;default:5"`  // 任务重试时间，单位为秒
	RetryAble   bool `json:"retryAble" gorm:"default:false"`       // 是否允许任务进行重试

	ExtraData JSON   `json:"extraData" gorm:"type:json"` // 扩展字段，用于存储外部服务调用时的信息
	Callback  string `json:"callback" gorm:"default:''"` // 外部请求的回调方式

	// 偏移检测相关
	CronDriftExpression   string     `json:"cronDriftExpression" gorm:"default:''"`      // 偏移检测任务的Cron表达式
	AutoRepairDrift       bool       `json:"autoRepairDrift" gorm:"default:false"`       // 是否进行自动纠偏
	OpenCronDrift         bool       `json:"openCronDrift" gorm:"default:false"`         // 是否开启偏移检测
	NextStartCronTaskTime *time.Time `json:"nextStartCronTaskTime" gorm:"type:datetime"` // 下次执行偏移检测任务的时间
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
	if err = sess.DropColumn(Env{}, "variables"); err != nil {
		return err
	}
	if err = sess.ModifyModelColumn(&Env{}, "triggers"); err != nil {
		return err
	}
	return nil
}

func (e *Env) DefaultStatPath() string {
	return path.Join(e.OrgId.String(), e.ProjectId.String(), e.Id.String(), "terraform.tfstate")
}

func (e *Env) MergeTaskStatus() string {
	if e.Deploying {
		e.Status = e.TaskStatus
	}
	return e.Status
}

type EnvDetail struct {
	Env
	Creator       string `json:"creator"`       // 创建人
	OperatorId    Id     `json:"operatorId"`    // 执行人ID
	Operator      string `json:"operator"`      // 执行人
	ResourceCount int    `json:"resourceCount"` // 资源数量
	TemplateName  string `json:"templateName"`  // 模板名称
	KeyName       string `json:"keyName"`       // 密钥名称
	TaskId        Id     `json:"taskId"`        // 当前作业ID
	CommitId      string `json:"commitId"`      // Commit ID
}
