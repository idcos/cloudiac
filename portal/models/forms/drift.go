package forms

import (
	"time"
)

// SearchEnvDriftsForm 漂移结果列表查询
type SearchEnvDriftsForm struct {
	NoPageSizeForm
	IsDrift   *bool      `json:"isDrift" form:"isDrift"`
	StartTime *time.Time `json:"startTime" form:"startTime" `
	EndTime   *time.Time `json:"endTime" form:"endTime" `
}
