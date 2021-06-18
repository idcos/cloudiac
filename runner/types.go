package runner

/*
portal 和 runner 通信使用的消息结构体
*/

// TaskStatusMessage runner 通知任务状态到 portal
// 增加StateListContent字段 用来下发terraform state list 命令输出的内容
type TaskStatusMessage struct {
	Status           string `json:"status" form:"status" `
	StatusCode       int    `json:"status_code" form:"status_code" `
	LogContent       []byte `json:"log_content" form:"log_content" `
	StateListContent []byte `json:"state_list_content" form:"state_list_content" `
	Code             string `json:"code" form:"code" `
	Error            string `json:"error" form:"error" `
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
