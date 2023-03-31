package apps

import (
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
	"fmt"
	"github.com/google/uuid"
	"net/http"
)

var (
	appStack = map[string]appInfo{
		"SaaS云管(融合Spot)": {
			OrgId:     "",
			ProjectId: "",
			TplId:     "",
		},
		"融合云虚拟机": {
			OrgId:     "org-ce8537o6vmfqkjkvei70",
			ProjectId: "p-ce864206vmfqkjkveumg",
			TplId:     "tpl-cgi1cf2t467v92vh972g",
		},
	}
)

type appInfo struct {
	OrgId     models.Id
	ProjectId models.Id
	TplId     models.Id
}

func DeclareEnv(c *ctx.ServiceContext, form *forms.DeclareEnvForm) (interface{}, e.Error) {
	if _, ok := appStack[form.AppStack]; !ok {
		return nil, e.New(e.BadParam)
	}

	// 初始化信息
	c.UserId = consts.SysUserId
	c.OrgId = appStack[form.AppStack].OrgId
	c.ProjectId = appStack[form.AppStack].ProjectId
	tplId := appStack[form.AppStack].TplId

	// 查询stack信息
	query := c.DB().Where("status = ?", models.Enable)
	tpl, err := services.GetTemplateById(query, tplId)
	if err != nil && err.Code() == e.TemplateNotExists {
		return nil, e.New(err.Code(), err, http.StatusBadRequest)
	} else if err != nil {
		c.Logger().Errorf("error get template, err %s", err)
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	}

	// 参数处理
	variables := make([]forms.Variable, 0)
	if form.AppStack == "融合云虚拟机" {
		if form.Instances.InstanceNumber != "" {
			variables = append(variables, forms.Variable{
				Scope: "env",
				Type:  "terraform",
				Name:  "instance_number",
				Value: form.Instances.InstanceNumber,
			})
		}
		if form.Instances.InstanceChargeType != "" {
			variables = append(variables, forms.Variable{
				Scope: "env",
				Type:  "terraform",
				Name:  "instance_charge_type",
				Value: form.Instances.InstanceChargeType,
			})
		}
		if form.Instances.SysDiskCategory != "" {
			variables = append(variables, forms.Variable{
				Scope: "env",
				Type:  "terraform",
				Name:  "system_disk_category",
				Value: form.Instances.SysDiskCategory,
			})
		}

		if form.Instances.SysDiskPerformance != "" {
			variables = append(variables, forms.Variable{
				Scope: "env",
				Type:  "terraform",
				Name:  "system_disk_performance_level",
				Value: form.Instances.SysDiskPerformance,
			})
		}

		if form.Instances.SysDiskSize != "" {
			variables = append(variables, forms.Variable{
				Scope: "env",
				Type:  "terraform",
				Name:  "system_disk_size",
				Value: form.Instances.SysDiskSize,
			})
		}

		if form.Instances.DataDiskSize != "" {
			variables = append(variables, forms.Variable{
				Scope: "env",
				Type:  "terraform",
				Name:  "data_disk_size",
				Value: form.Instances.DataDiskSize,
			})
		}
		if form.Instances.DataDiskCategory != "" {
			variables = append(variables, forms.Variable{
				Scope: "env",
				Type:  "terraform",
				Name:  "disk_category",
				Value: form.Instances.DataDiskCategory,
			})
		}
		if form.Instances.DataDiskPerformance != "" {
			variables = append(variables, forms.Variable{
				Scope: "env",
				Type:  "terraform",
				Name:  "performance_level",
				Value: form.Instances.DataDiskPerformance,
			})
		}
		if form.Instances.InstanceType != "" {
			variables = append(variables, forms.Variable{
				Scope: "env",
				Type:  "terraform",
				Name:  "instance_type",
				Value: form.Instances.InstanceType,
			})
		}
		if form.Instances.UserData != "" {
			variables = append(variables, forms.Variable{
				Scope: "env",
				Type:  "terraform",
				Name:  "user_data",
				Value: form.Instances.UserData,
			})
		}
		if form.Instances.Tags != "" {
			variables = append(variables, forms.Variable{
				Scope: "env",
				Type:  "terraform",
				Name:  "tags",
				Value: form.Instances.Tags,
			})
		}
		if form.Instances.FirstIndex != "" {
			variables = append(variables, forms.Variable{
				Scope: "env",
				Type:  "terraform",
				Name:  "first_index",
				Value: form.Instances.FirstIndex,
			})
		}
		if form.Instances.EnvironmentId != "" {
			variables = append(variables, forms.Variable{
				Scope: "env",
				Type:  "terraform",
				Name:  "environment_id",
				Value: form.Instances.EnvironmentId,
			})
		}
		if form.Instances.KeyName != "" {
			variables = append(variables, forms.Variable{
				Scope: "env",
				Type:  "terraform",
				Name:  "key_name",
				Value: form.Instances.KeyName,
			})
		}
		if form.Instances.ImageId != "" {
			variables = append(variables, forms.Variable{
				Scope: "env",
				Type:  "terraform",
				Name:  "image_id",
				Value: form.Instances.ImageId,
			})
		}
	}

	// 构建环境参数
	f := &forms.CreateEnvForm{
		TplId:        tpl.Id,
		Name:         fmt.Sprintf("%s-%s", form.AppStack, uuid.New().String()),
		AutoApproval: true,
		TaskType:     "apply",
		Variables:    variables,

		// 模板参数
		TfVarsFile:   tpl.TfVarsFile,
		PlayVarsFile: tpl.PlayVarsFile,
		Playbook:     tpl.Playbook,
		Revision:     tpl.RepoRevision,
		KeyId:        tpl.KeyId,
		Workdir:      tpl.Workdir,

		ExtraData: form.ExtraData,
		Source:    "CPG",
	}

	return CreateEnv(c, f)
}
