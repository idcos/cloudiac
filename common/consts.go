package common

const (
	TaskTypePlan    = "plan"    // 计划执行，不会修改资源或做服务配置
	TaskTypeApply   = "apply"   // 执行 terraform apply 和 playbook
	TaskTypeDestroy = "destroy" // 销毁，删除所有资源

	TaskPending  = "pending"
	TaskRunning  = "running"
	TaskFailed   = "failed"
	TaskComplete = "complete"
	TaskTimeout  = "timeout"

	TaskStepInit    = "init"
	TaskStepPlan    = "plan"
	TaskStepApply   = "apply"
	TaskStepDestroy = "destroy"
	TaskStepPlay    = "play"    // play playbook
	TaskStepCommand = "command" // run command

	TaskStepPending  = "pending"
	TaskStepRunning  = "running"
	TaskStepFailed   = "failed"
	TaskStepComplete = "complete"
	TaskStepTimeout  = "timeout"
)
