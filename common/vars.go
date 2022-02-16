// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

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
