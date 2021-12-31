// Copyright 2021 CloudJ Company Limited. All rights reserved.

package common

var (
	DemoOrgId string

	TaskJobTypes = []string{
		TaskJobPlan,
		TaskJobApply,
		TaskJobDestroy,
		// 0.3
		TaskJobScan,
		TaskJobParse,
		// 0.4
		TaskJobEnvScan,
		TaskJobEnvParse,
		TaskJobTplScan,
		TaskJobTplParse,
	}
)
