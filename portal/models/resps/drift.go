package resps

import "cloudiac/portal/models"

type ResourceDriftResp struct {
	models.Resource
	DriftDetail string `json:"driftDetail"`
}
