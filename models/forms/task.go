package forms

type CreateTaskForm struct {
	BaseForm

	TaskName     string `json:"taskName" form:"taskName" `
	RunnerIp     string `json:"runnerIp" form:"runnerIp" `
	RunnerPort   string `json:"runnerPort" form:"runnerPort" `
	TemplateId   uint64 `json:"templateId" form:"templateId" `
	TemplateGuid string `json:"templateGuid" form:"templateGuid" `
	TaskType     string `json:"taskType" form:"taskType" `
	Creator      uint   `json:"creator" grom:"not null;comment:'创建人'"`
}

type DetailTaskForm struct {
	BaseForm
	TaskId uint64 `json:"taskId" form:"taskId" `
}
