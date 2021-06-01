package runner

/*
portal 和 runner 通信使用的消息结构体
*/

// TaskStatusMessage runner 通知任务状态到 portal
type TaskStatusMessage struct {
	Status     string `json:"status" form:"status" `
	StatusCode int    `json:"status_code" form:"status_code" `
	LogContent []byte `json:"log_content" form:"log_content" `
	Code       string `json:"code" form:"code" `
	Error      string `json:"error" form:"error" `
}

type ErrorMessage struct {
	Error string `json:"error"`
}
