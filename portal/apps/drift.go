package apps

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/libs/page"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/models/resps"
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

// EnvDriftSearch 环境漂移结果查询
func EnvDriftSearch(c *ctx.ServiceContext, envId models.Id, form *forms.SearchEnvDriftsForm) (*page.PageResp, e.Error) {
	query := services.QueryTaskDrift(c.DB())
	query.Where("env_id = ?", envId)
	if form.Keyword != "" {
		// TODO 关键字查询,需要先确认查询啥
		//query.Where("name")
	}
	if form.IsDrift != nil {
		query.Where("is_drift = ?", form.IsDrift)
	}
	query.Order("created_at DESC")
	p := page.New(form.CurrentPage(), form.PageSize(), query)
	drifts := make([]*resps.TaskDriftResp, 0)
	if err := p.Scan(&drifts); err != nil {
		return nil, e.New(e.DBError, err)
	}
	return &page.PageResp{
		Total:    p.MustTotal(),
		PageSize: p.Size,
		List:     drifts,
	}, nil
}

// EnvDriftResourceSearch 查询偏移资源信息
func EnvDriftResourceSearch(c *ctx.ServiceContext, envId models.Id, taskId models.Id) ([]*resps.ResourceDriftResp, e.Error) {
	query := services.QueryResourceDrift(c.DB())
	query.Where("iac_resource.env_id = ? and rd.task_id = ?", envId, taskId)
	rdr := make([]*resps.ResourceDriftResp, 0)
	if err := query.Scan(&rdr); err != nil {
		return nil, e.New(e.DBError, err)
	}
	return rdr, nil
}
