package apps

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models"
	"cloudiac/portal/services"
)

// EnvDriftDetail 环境漂移详情
func EnvDriftDetail(c *ctx.ServiceContext, envId models.Id) (*models.EnvDrift, e.Error) {
	query := services.QueryEnvDetail(c.DB(), c.OrgId, c.ProjectId)

	envDetail, err := services.GetEnvDetailById(query, envId)
	if err != nil {
		return nil, err
	}
	return &models.EnvDrift{
		EnvId:            envDetail.Id,
		IsDrift:          envDetail.IsDrift,
		CronDriftExpress: envDetail.CronDriftExpress,
		AutoRepairDrift:  envDetail.AutoRepairDrift,
		OpenCronDrift:    envDetail.OpenCronDrift,
	}, nil
}
