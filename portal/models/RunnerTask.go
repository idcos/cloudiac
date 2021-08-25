package models

type BaseTask struct {
	SoftDeleteModel

	/* 通用任务参数 */
	Flow     TaskFlow `json:"-" gorm:"type:text"`        // 执行流程
	CurrStep int      `json:"currStep" gorm:"default:0"` // 当前在执行的流程步骤

	// 任务每一步的执行超时(整个任务无超时控制)
	StepTimeout int `json:"stepTimeout" gorm:"default:600;comment:执行超时"`

	Status  string `json:"status" gorm:"type:enum('pending','running','approving','rejected','failed','complete','timeout');default:'pending'" enums:"'pending','running','approving','rejected','failed','complete','timeout'"`
	Message string `json:"message"` // 任务的状态描述信息，如失败原因等

	StartAt *Time `json:"startAt" gorm:"type:datetime;comment:任务开始时间"` // 任务开始时间
	EndAt   *Time `json:"endAt" gorm:"type:datetime;comment:任务结束时间"`   // 任务结束时间
}

// ScanTask 合规扫描任务
type ScanTask struct {
	BaseTask

	// 任务参数
	OrgId     Id `json:"orgId" gorm:"size:32;not null"`     // 组织ID
	ProjectId Id `json:"projectId" gorm:"size:32;not null"` // 项目ID
	TplId     Id `json:"TplId" gorm:"size:32;not null"`     // 模板ID
	EnvId     Id `json:"envId" gorm:"size:32;not null"`     // 环境ID

	Name      string `json:"name" gorm:"not null;comment:任务名称"` // 任务名称
	CreatorId Id     `json:"creatorId" gorm:"size:32;not null"` // 创建人ID

	Type string `json:"type" gorm:"not null;enum('plan','apply','destroy','scan'')" enums:"'plan','apply','destroy','scan'"` // 任务类型。1. plan: 计划 2. apply: 部署 3. destroy: 销毁

	RepoAddr string `json:"repoAddr" gorm:"not null"`
	Revision string `json:"revision" gorm:"not null"`
	CommitId string `json:"commitId" gorm:"not null"` // 创建任务时 revision 对应的 commit id

	Workdir      string   `json:"workdir" gorm:"default:''"`
	Playbook     string   `json:"playbook" gorm:"default:''"`
	TfVarsFile   string   `json:"tfVarsFile" gorm:"default:''"`
	PlayVarsFile string   `json:"playVarsFile" gorm:"default:''"`
	Targets      StrSlice `json:"targets" gorm:"type:json"` // 指定 terraform target 参数

	Variables TaskVariables `json:"variables" gorm:"type:json"` // 本次执行使用的所有变量(继承、覆盖计算之后的)

	StatePath string `json:"statePath" gorm:"not null"`

	// 扩展属性，包括 source, transitionId 等
	Extra TaskExtra `json:"extra" gorm:"type:json"` // 扩展属性

	KeyId           Id     `json:"keyId" gorm:"size32"`      // 部署密钥ID
	RunnerId        string `json:"runnerId" gorm:"not null"` // 部署通道
	AutoApprove     bool   `json:"autoApproval" gorm:"default:false"`
	StopOnViolation bool   `json:"stopOnViolation" gorm:"default:false"`

	// 任务执行结果，如 add/change/delete 的资源数量、outputs 等
	Result TaskResult `json:"result" gorm:"type:json"` // 任务执行结果
}

func (ScanTask) TableName() string {
	return "iac_scan_task"
}
