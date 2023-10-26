package resps

import "cloudiac/portal/models"

type TaskDriftResp struct {
	models.TaskDriftInfo
}
type ResourceDriftResp struct {
	models.Resource
	DriftDetail string `json:"driftDetail"`
}
