package runner

/*
portal 和 runner 通信使用的消息结构体
*/

type EnvVariables struct {
}

type TaskEnv struct {
	Id           string `json:"id" binding:"required"`
	Workdir      string `json:"workdir"`
	TfVarsFile   string `json:"tfVarsFile"`
	Playbook     string `json:"playbook"`
	PlayVarsFile string `json:"playVarsFile"`

	EnvironmentVars map[string]string `json:"environment"`
	TerraformVars   map[string]string `json:"terraform"`
	AnsibleVars     map[string]string `json:"ansible"`
}

type StateStore struct {
	Backend string `json:"backend" binding:"required"`
	Scheme  string `json:"scheme" binding:"required"`
	Path    string `json:"path" binding:"required"`
	Address string `json:"address" binding:"required"`
}

type RunTaskReq struct {
	Env          TaskEnv    `json:"env" binding:"required"`
	RunnerId     string     `json:"runnerId" binding:""`
	TaskId       string     `json:"taskId" binding:"required"`
	Step         int        `json:"step" binding:""`
	StepType     string     `json:"stepType" binding:"required"`
	StepArgs     []string   `json:"stepArgs"`
	DockerImage  string     `json:"dockerImage"`
	StateStore   StateStore `json:"stateStore" binding:"required"`
	RepoAddress  string     `json:"repoAddress" binding:"required"` // 带 token 的完整路径
	RepoRevision string     `json:"repoRevision" binding:"required"`

	Timeout    int    `json:"timeout"`
	PrivateKey string `json:"privateKey"`
}

type TaskStatusReq struct {
	EnvId  string `json:"envId" form:"envId" binding:"required"`
	TaskId string `json:"taskId" form:"taskId" binding:"required"`
	Step   int    `json:"step" form:"step" binding:""`
}

type TaskLogReq TaskStatusReq

// TaskStatusMessage runner 通知任务状态到 portal
type TaskStatusMessage struct {
	Exited   bool `json:"exited"`
	ExitCode int  `json:"status_code"`

	LogContent  []byte `json:"logContent"`
	TfStateJson []byte `json:"tfStateJson"`
}

type ErrorMessage struct {
	Error string `json:"error"`
}

type Response struct {
	Error  string      `json:"error,omitempty"`
	Result interface{} `json:"result"`
}
