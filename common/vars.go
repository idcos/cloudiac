// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package common

var (
	// 旧版本中演示组织是全局使用一个，会在初始化时设置该变量。
	// 新版本演示组织是每个用户有一个，所以该特性不再需要。
	// DemoOrgId string

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
