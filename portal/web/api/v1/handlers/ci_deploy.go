package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
)

func CiDeployEnv(c *ctx.GinRequest) {
	// 创建环境部署环境，没有项目创建项目
	form := &forms.CreateCiDeployForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.CreateCiDeployEnvs(c.Service(), form))

}
