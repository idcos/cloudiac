package runner

/*
portal 和 runner 通信使用的消息结构体
*/

// TaskStatusMessage runner 通知任务状态到 portal
type TaskStatusMessage struct {
	Exited   bool `json:"exited"`
	ExitCode int  `json:"status_code"`

	LogContent       []byte `json:"log_content"`
	StateListContent []byte `json:"state_list_content"`
}

type ErrorMessage struct {
	Error string `json:"error"`
}

type StateStore struct {
	SaveState           bool   `json:"save_state"`
	Backend             string `json:"backend" default:"consul"`
	Scheme              string `json:"scheme" default:"http"`
	StateKey            string `json:"state_key"`
	StateBackendAddress string `json:"state_backend_address"`
	Lock                bool   `json:"lock"`
}
