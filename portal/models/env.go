// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package models

import (
	"cloudiac/common"
	"cloudiac/portal/libs/db"
	"path"
	"time"
)

const (
	EnvStatusActive    = "active"    // 成功部署
	EnvStatusFailed    = "failed"    // apply 过程中出现错误
	EnvStatusInactive  = "inactive"  // 资源未部署
	EnvStatusDestroyed = "destroyed" // 已销毁

	//EnvStatusDeploying = "deploying" // apply 运行中(plan 作业不改变状态)
	//EnvStatusApproving = "approving" // 等待审批
)

var (
	EnvStatus     = []string{EnvStatusActive, EnvStatusFailed, EnvStatusInactive, EnvStatusDestroyed}
	EnvTaskStatus = []string{TaskRunning, TaskApproving} // 环境 taskStatus 有效值
)

type Env struct {
	SoftDeleteModel
	OrgId     Id `json:"orgId" gorm:"size:32;not null"`                                           // 组织ID
	ProjectId Id `json:"projectId" gorm:"size:32;not null"`                                       // 项目ID
	TplId     Id `json:"tplId" gorm:"size:32;not null"`                                           // 模板ID
	CreatorId Id `json:"creatorId" gorm:"size:32;not null"`                                       // 创建人ID
	TokenId   Id `json:"tokenId" gorm:"size:32;comment:tokenId" example:"t-cgptjsit467j7gq5jiv0"` // Token ID

	Name        string `json:"name" gorm:"not null"`                                             // 环境名称
	Description Text   `json:"description" gorm:"type:text"`                                     // 环境描述
	Status      string `json:"status" gorm:"" enums:"'active','failed','inactive', 'destroyed'"` // 环境状态, active活跃, inactive非活跃,failed错误,running部署中,approving审批中
	// 任务状态，只同步部署任务的状态(apply,destroy)，plan 任务不会对环境产生影响，所以不同步
	TaskStatus  string `json:"taskStatus" gorm:"default:''"`                 // type:enum('','approving','running')
	Archived    bool   `json:"archived" gorm:"default:false"`                // 是否已归档
	StepTimeout int    `json:"stepTimeout" gorm:"default:3600;comment:部署超时"` // 步骤超时时间（单位：秒）
	OneTime     bool   `json:"oneTime" gorm:"default:false"`                 // 一次性环境标识
	Deploying   bool   `json:"deploying" gorm:"not null;default:false"`      // 是否正在执行部署

	Tags Text `json:"tags" gorm:"type:text"`

	StatePath string `json:"statePath" gorm:"not null" swaggerignore:"true"` // Terraform tfstate 文件路径（内部）

	// 环境可以覆盖模板中的 vars file 配置，具体说明见 Template model
	TfVarsFile   string `json:"tfVarsFile" gorm:"default:''"`   // Terraform tfvars 变量文件路径
	PlayVarsFile string `json:"playVarsFile" gorm:"default:''"` // Ansible 变量文件路径
	Playbook     string `json:"playbook" gorm:"default:''"`     // Ansible playbook 入口文件路径

	// 任务相关参数，获取详情的时候，如果有 last_task_id 则返回 last_task_id 相关参数
	RunnerId   string `json:"runnerId" gorm:"size:32"`            //部署通道ID
	RunnerTags string `json:"runnerTags" gorm:"size:256"`         //部署通道Tags,逗号分割
	Revision   string `json:"revision" gorm:"size:64;default:''"` // Vcs仓库分支/标签
	KeyId      Id     `json:"keyId" gorm:"size:32"`               // 部署密钥ID
	Workdir    string `json:"workdir" gorm:"size:32;default:''"`  // 工作目录

	LastTaskId    Id `json:"lastTaskId" gorm:"size:32"`          // 最后一次部署或销毁任务的 id(plan 任务不记录)
	LastResTaskId Id `json:"lastResTaskId" gorm:"index;size:32"` // 最后一次进行了资源列表统计的部署任务的 id

	LastScanTaskId Id `json:"lastScanTaskId" gorm:"size:32"` // 最后一次策略扫描任务 id

	AutoApproval    bool `json:"autoApproval" gorm:"default:false"`    // 是否自动审批
	StopOnViolation bool `json:"stopOnViolation" gorm:"default:false"` // 当合规不通过是否中止部署

	TTL           string `json:"ttl" gorm:"default:'0'" example:"1h/1d"` // 生命周期
	AutoDestroyAt *Time  `json:"autoDestroyAt" gorm:"type:datetime"`     // 自动销毁时间

	// 该 id 在创建自动销毁任务后保存，并在销毁任务执行完成后清除
	AutoDestroyTaskId Id `json:"-"  gorm:"default:''"` // 自动销毁任务 id

	// 触发器设置
	Triggers StringArray `json:"triggers" gorm:"type:text" swaggertype:"array,string"` // 触发器。commit（每次推送自动部署），prmr（提交PR/MR的时候自动执行plan）

	// 任务重试
	RetryNumber int  `json:"retryNumber" gorm:"size:32;default:3"` // 任务重试次数
	RetryDelay  int  `json:"retryDelay" gorm:"size:32;default:5"`  // 任务重试时间，单位为秒
	RetryAble   bool `json:"retryAble" gorm:"default:false"`       // 是否允许任务进行重试

	ExtraData JSON   `json:"extraData" gorm:"type:text"` // 扩展字段，用于存储外部服务调用时的信息
	Callback  string `json:"callback" gorm:"default:''"` // 外部请求的回调方式

	// 偏移检测相关
	CronDriftExpress  string     `json:"cronDriftExpress" gorm:"default:''"`     // 偏移检测任务的Cron表达式
	AutoRepairDrift   bool       `json:"autoRepairDrift" gorm:"default:false"`   // 是否进行自动纠偏
	OpenCronDrift     bool       `json:"openCronDrift" gorm:"default:false"`     // 是否开启偏移检测
	NextDriftTaskTime *time.Time `json:"nextDriftTaskTime" gorm:"type:datetime"` // 下次执行偏移检测任务的时间

	// 合规相关
	PolicyEnable bool `json:"policyEnable" gorm:"default:false"` // 是否开启合规检测

	//环境锁定
	Locked bool `json:"locked" gorm:"default:false"`

	IsDemo bool `json:"isDemo" gorm:"default:false"` // 是否是演示环境

	Targets StrSlice `json:"targets,omitempty" gorm:"type:text"` // 指定部署的资源
	// 自动部署相关
	AutoDeployCron   string `json:"autoDeployCron" gorm:"default:''"`  // 自动部署任务的Cron表达式
	AutoDeployAt     *Time  `json:"autoDeployAt" gorm:"type:datetime"` // 下次执行自动部署任务的时间
	AutoDeployTaskId Id     `json:"-"  gorm:"default:''"`              // 自动部署任务 id

	// 自动销毁相关
	AutoDestroyCron string `json:"autoDestroyCron" gorm:"default:''"` // 自动销毁任务的Cron表达式
	// 下次执行自动部署任务的时间 和 自动部署任务id 复用之前的
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
	if err = sess.ModifyModelColumn(&Env{}, "status"); err != nil {
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
	TokenName     string `json:"tokenName"`     // Token 名称
	ResourceCount int    `json:"resourceCount"` // 资源数量
	TemplateName  string `json:"templateName"`  // 模板名称
	KeyName       string `json:"keyName"`       // 密钥名称
	TaskId        Id     `json:"taskId"`        // 当前作业ID
	CommitId      string `json:"commitId"`      // Commit ID
	IsDrift       bool   `json:"isDrift"`
	PolicyEnable  bool   `json:"policyEnable"` // 是否开启合规检测
	PolicyStatus  string `json:"policyStatus"` // 环境合规检测任务状态

	// PolicyGroup 必须配置 struct tag `gorm:"-"`。
	// 因为我们定义了 model struct PolicyGroup，
	// gorm 解析该结构体的 PolicyGroup 字段时会将其理解为 PolicyGroup model 的关联字段，
	// 但解析类型却发现是一个 []string， 而非 []struct{}，导致报错  "[error] unsupported data type: &[]"
	// (这个报错只在 gorm 日志中打印，db.Error 无错误)。
	PolicyGroup []string `json:"policyGroup" gorm:"-"` // 环境相关合规策略组
	RunnerTags  []string `json:"runnerTags" gorm:"-"`  // 将其转为数组返回给前端

	MonthCost float32 `json:"monthCost"` // 环境月度成本
	IsBilling bool    `json:"isBilling"` // 是否开启账单采集
	EnvTags   []Tag   `json:"envTags" form:"envTags" gorm:"-"`
	UserTags  []Tag   `json:"userTags" form:"userTags" gorm:"-"`
}

func (c *EnvDetail) UpdateEnvPolicyStatus() {
	if c.PolicyEnable {
		if c.PolicyStatus == "failed" {
			c.PolicyStatus = common.PolicyStatusViolated
		} else if c.PolicyStatus == "" {
			c.PolicyStatus = common.PolicyStatusEnable
		}
	} else {
		c.PolicyStatus = common.PolicyStatusDisable
	}
}

type Tag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
