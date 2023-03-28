package apps

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
	"github.com/google/uuid"
	"net/http"
)

func GcpDeploy(c *ctx.ServiceContext, form *forms.GcpDeployForm) (interface{}, e.Error) {
	c.OrgId = "org-cb39pjflgt4e4oeql6i0"
	c.ProjectId = "p-cb39q6vlgt4e4oeql6j0"
	tplId := "tpl-cet7voit467gcjghqbp0"
	query := c.DB().Where("status = ?", models.Enable)
	tpl, err := services.GetTemplateById(query, models.Id(tplId))
	if err != nil && err.Code() == e.TemplateNotExists {
		return nil, e.New(err.Code(), err, http.StatusBadRequest)
	} else if err != nil {
		c.Logger().Errorf("error get template, err %s", err)
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	}

	key := uuid.New().String()

	createSpotInstance := "false"
	if form.ChargeType == "spot" {
		createSpotInstance = "true"
	}
	f := &forms.CreateEnvForm{
		TplId:        tpl.Id,
		Name:         key,
		AutoApproval: true,
		TaskType:     "apply",
		Variables: []forms.Variable{
			forms.Variable{
				Scope: "env",
				Type:  "terraform",
				Name:  "zone_id",
				Value: form.ZoneId,
			},
			forms.Variable{
				Scope: "env",
				Type:  "terraform",
				Name:  "create_spot_instance",
				Value: createSpotInstance,
			},
		},

		// 模板参数
		TfVarsFile:   tpl.TfVarsFile,
		PlayVarsFile: tpl.PlayVarsFile,
		Playbook:     tpl.Playbook,
		Revision:     tpl.RepoRevision,
		KeyId:        tpl.KeyId,
		Workdir:      tpl.Workdir,

		ExtraData:   models.JSON(form.ExtraData),
		Source:      "GCP",
	}
	return CreateEnv(c, f)
}
