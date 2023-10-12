package resps

import "cloudiac/portal/models"

type TaskDriftResp struct {
	models.TaskDrift
	Status string `json:"status"` // 漂移任务结果
}
type ResourceDriftResp struct {
	models.Resource
	DriftDetail string `json:"driftDetail"`
}
