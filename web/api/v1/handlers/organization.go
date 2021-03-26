package handlers

import (
	"cloudiac/apps"
	"cloudiac/consts/e"
	"cloudiac/libs/ctrl"
	"cloudiac/libs/ctx"
	"cloudiac/models/forms"
)

type Organization struct {
	ctrl.BaseController
}

func (Organization) Create(c *ctx.GinRequestCtx) {
	form := &forms.CreateOrganizationForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.CreateOrganization(c.ServiceCtx(), form))
}

func (Organization) Search(c *ctx.GinRequestCtx) {
	form := forms.SearchOrganizationForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.SearchOrganization(c.ServiceCtx(), &form))
}

func (Organization) Update(c *ctx.GinRequestCtx) {
	form := forms.UpdateOrganizationForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateOrganization(c.ServiceCtx(), &form))
}

func (Organization) Delete(c *ctx.GinRequestCtx) {
	//form := forms.DeleteUserForm{}
	//if err := c.Bind(&form); err != nil {
	//	return
	//}
	//c.JSONResult(apps.DeleteUser(c.ServiceCtx(), &form))
	c.JSONError(e.New(e.NotImplement))
}

func (Organization) Detail(c *ctx.GinRequestCtx) {
	form := forms.DetailOrganizationForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.OrganizationDetail(c.ServiceCtx(), &form))
}

func (Organization) ChangeOrgStatus(c *ctx.GinRequestCtx) {
	form := forms.DisableOrganizationForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.ChangeOrgStatus(c.ServiceCtx(), &form))
}

func (Organization) ListRepos(c *ctx.GinRequestCtx) {
	orgId, _ := c.QueryInt("orgId")
	searchKey := c.Query("searchKey")
	c.JSONResult(apps.ListOrganizationRepos(c.ServiceCtx(), orgId, searchKey))
}

func (Organization) ListBranches(c *ctx.GinRequestCtx) {
	orgId, _ := c.QueryInt("orgId")
	repoId, _ := c.QueryInt("repoId")
	c.JSONResult(apps.ListRepositoryBranches(c.ServiceCtx(), orgId, repoId))
}

func (Organization) GetReadmeContent(c *ctx.GinRequestCtx) {
	orgId, _ := c.QueryInt("orgId")
	repoId, _ := c.QueryInt("repoId")
	branch := c.Query("branch")
	if branch == "" {
		branch = "master"
	}
	c.JSONResult(apps.GetReadmeContent(c.ServiceCtx(), orgId, repoId, branch))
}

func (Organization) ListNotificationCfgs(c *ctx.GinRequestCtx) {
	orgId, _ := c.QueryInt("orgId")
	c.JSONResult(apps.ListNotificationCfgs(c.ServiceCtx(), orgId))
}

func (Organization) CreateNotificationCfgs(c *ctx.GinRequestCtx) {
	form := &forms.CreateNotificationCfgForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.CreateNotificationCfg(c.ServiceCtx(), form))
}

func (Organization) DeleteNotificationCfgs(c *ctx.GinRequestCtx) {
	cfgId, _ := c.QueryInt("notificationId")
	c.JSONResult(apps.DeleteNotificationCfg(c.ServiceCtx(), cfgId))
}

func (Organization) UpdateNotificationCfgs(c *ctx.GinRequestCtx) {
	form := &forms.UpdateNotificationCfgForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateNotificationCfg(c.ServiceCtx(), form))
}
