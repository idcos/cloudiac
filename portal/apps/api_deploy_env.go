package apps

import (
	"cloudiac/common"
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
)

type CiDeployResp struct {
	EnvId      models.Id `json:"envId"`
	TaskId     models.Id `json:"taskId"`
	EnvStatus  string    `json:"envStatus"`
	TaskStatus string    `json:"taskStatus"`
}

func createProjectByApiToken(c *ctx.ServiceContext, form *forms.CreateCiDeployForm) (*models.Project, e.Error) {
	createProjectForm := &forms.CreateProjectForm{
		Name: form.ProjectName,
	}
	p, err := CreateProject(c, createProjectForm)
	if err != nil {
		return nil, err
	}
	return p.(*models.Project), err
}

func createEnvByApiToken(c *ctx.ServiceContext, form *forms.CreateCiDeployForm) (*models.EnvDetail, e.Error) {
	// 未找到环境进行环境创建
	createEnvForm := &forms.CreateEnvForm{
		TplId:        form.TplId,
		KeyId:        form.KeyId,
		TfVarsFile:   form.TfVarsFile,
		PlayVarsFile: form.Playbook,
		TaskType:     common.TaskTypeApply,
		AutoApproval: true,
		Name:         form.EnvName,
	}
	envVars := []forms.Variables{}
	for key, value := range form.TerraformVars {
		envVar := forms.Variables{
			Scope: consts.ScopeEnv,
			Type:  consts.VarTypeTerraform,
			Name:  key,
			Value: value,
		}
		envVars = append(envVars, envVar)
	}
	for key, value := range form.EnvVars {
		envVar := forms.Variables{
			Scope: consts.ScopeEnv,
			Type:  consts.VarTypeEnv,
			Name:  key,
			Value: value,
		}
		envVars = append(envVars, envVar)
	}

	createEnvForm.TTL = form.TTL
	createEnvForm.DestroyAt = form.DestroyAt
	createEnvForm.Variables = envVars
	envDetail, err := CreateEnv(c, createEnvForm)
	if err != nil {
		return nil, e.New(e.DBError, err)
	}
	return envDetail, nil
}

func searchEnv(c *ctx.ServiceContext, envName string) (*models.Env, e.Error) {
	query := c.DB()
	env := &models.Env{}
	if err := query.Where("name = ? and org_id = ? AND project_id = ?", envName, c.OrgId, c.ProjectId).First(env); err != nil {
		if e.IsRecordNotFound(err) {
			return env, nil
		} else {
			return nil, e.New(e.DBError, err)
		}
	}
	return env, nil
}

func CreateCiDeployEnvs(c *ctx.ServiceContext, form *forms.CreateCiDeployForm) (interface{}, e.Error) {
	query := c.DB()
	project := &models.Project{}
	// 查询项目是否存在，不存在就创建
	if err := query.Where("name = ?", form.ProjectName).First(project); err != nil {
		if e.IsRecordNotFound(err) {
			project, err = createProjectByApiToken(c, form)
		} else {
			return nil, e.New(e.DBError, err)
		}
	}
	// 查询环境是否存在，存在重新部署，否则创建新环境
	c.ProjectId = project.Id
	env, err := searchEnv(c, form.EnvName)
	if err != nil {
		return nil, err
	}
	envDetail := &models.EnvDetail{}
	if env.Name == "" {
		envDetail, err = createEnvByApiToken(c, form)
		if err != nil {
			return nil, err
		}
	} else {
		// 环境存在，调用deploy重新部署
		searchVarForm := &forms.SearchVariableForm{
			TplId: form.TplId,
			EnvId: env.Id,
			Scope: consts.ScopeEnv,
		}
		deployVar, err := SearchVariable(c, searchVarForm)
		if err != nil {
			return nil, err
		}
		deployForm := &forms.DeployEnvForm{
			Id:         env.Id,
			Name:       form.EnvName,
			Playbook:   form.Playbook,
			Revision:   env.Revision,
			TaskType:   common.TaskTypeApply,
			TfVarsFile: form.TfVarsFile,
			Triggers:   env.Triggers,
		}
		for _, v := range deployVar.([]VariableResp) {
			tempVar := forms.Variables{
				Id:          v.Id,
				Scope:       v.Scope,
				Type:        v.Type,
				Name:        v.Name,
				Value:       v.Value,
				Sensitive:   v.Sensitive,
				Description: v.Description,
				Options:     v.Options,
			}
			deployForm.Variables = append(deployForm.Variables, tempVar)
		}

		envDetail, err = EnvDeploy(c, deployForm)
		if err != nil {
			return nil, err
		}
	}
	task, err := services.GetTaskById(c.DB(), envDetail.LastTaskId)
	if err != nil {
		return nil, err
	}
	return CiDeployResp{
		EnvId:     envDetail.Id,
		TaskId:    envDetail.LastTaskId,
		EnvStatus: envDetail.Status,
		// 需要固定写死，同步直接去任务查询，此时任务iac_storege没有完成写入，任务一定不存在
		TaskStatus: task.Status,
	}, nil

}
