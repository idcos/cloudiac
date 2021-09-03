package apps

import (
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
	"fmt"
	"net/http"
)

// CreatePolicyRel 创建策略关系
func CreatePolicyRel(c *ctx.ServiceContext, form *forms.CreatePolicyRelForm) ([]models.PolicyRel, e.Error) {
	c.AddLogField("action", fmt.Sprintf("create policy group relation %s%s", form.EnvId, form.TplId))

	var (
		env  *models.Env
		tpl  *models.Template
		rels []models.PolicyRel
		err  e.Error
	)
	tx := c.Tx()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	if (!form.HasKey("envId") && !form.HasKey("tplId")) ||
		(form.HasKey("envId") && form.HasKey("tplId")) ||
		len(form.PolicyGroupIds) == 0 {
		_ = tx.Rollback()
		return nil, e.New(e.BadParam, http.StatusBadRequest)
	}

	if form.HasKey("envId") {
		env, err = services.GetEnvById(tx, form.EnvId)
		if err != nil {
			_ = tx.Rollback()
			return nil, e.New(err.Code(), err, http.StatusBadRequest)
		}
	}
	if form.HasKey("tplId") {
		tpl, err = services.GetTemplateById(tx, form.TplId)
		if err != nil {
			_ = tx.Rollback()
			return nil, e.New(err.Code(), err, http.StatusBadRequest)
		}
	}

	for _, groupId := range form.PolicyGroupIds {
		group, err := services.GetPolicyGroupById(tx, groupId)
		if err != nil {
			_ = tx.Rollback()
			return nil, e.New(err.Code(), err, http.StatusBadRequest)
		}
		if env != nil {
			rels = append(rels, models.PolicyRel{
				OrgId:     env.OrgId,
				ProjectId: env.ProjectId,
				GroupId:   group.Id,
				EnvId:     env.Id,
				Scope:     consts.ScopeEnv,
			})
		} else {
			rels = append(rels, models.PolicyRel{
				OrgId:   tpl.OrgId,
				GroupId: group.Id,
				TplId:   tpl.Id,
				Scope:   "template",
			})
		}
	}

	if er := models.CreateBatch(tx, rels); er != nil {
		_ = tx.Rollback()
		return nil, e.New(e.DBError, er)
	}

	if err := tx.Commit(); err != nil {
		c.Logger().Errorf("error commit policy relations, err %s", err)
		_ = tx.Rollback()
		return nil, e.New(e.DBError, err)
	}

	return rels, nil
}
