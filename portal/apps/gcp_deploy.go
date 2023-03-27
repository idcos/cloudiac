package apps

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
)

func GcpDeploy(c *ctx.ServiceContext) (interface{}, e.Error) {
	c.OrgId = ""
	c.ProjectId = ""
	f := &forms.CreateEnvForm{
		TplId:            "",
		Name:             "",
		Tags:             "",
		AutoApproval:     true,
		TaskType:         "",
		Targets:          "",
		RunnerId:         "",
		RunnerTags:       nil,
		Revision:         "",
		StepTimeout:      0,
		Variables:        nil,
		TfVarsFile:       "",
		PlayVarsFile:     "",
		Playbook:         "",
		KeyId:            "",
		Workdir:          "",
		RetryNumber:      0,
		RetryDelay:       0,
		RetryAble:        false,
		ExtraData:        nil,
		VarGroupIds:      nil,
		DelVarGroupIds:   nil,
		SampleVariables:  nil,
		Callback:         "",
		CronDriftExpress: "",
		AutoRepairDrift:  false,
		OpenCronDrift:    false,
		PolicyEnable:     false,
		PolicyGroup:      nil,
		Source:           "",
		AutoDeployCron:   "",
		AutoDestroyCron:  "",
	}
	return CreateEnv(c, f)
}
