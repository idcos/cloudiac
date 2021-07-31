// Copyright 2021 CloudJ Company Limited. All rights reserved.

package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
)

type ResourceAccount struct {
	ctrl.GinController
}

func (ResourceAccount) Search(c *ctx.GinRequest) {
	form := &forms.SearchResourceAccountForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.SearchResourceAccount(c.Service(), form))
}

func (ResourceAccount) Create(c *ctx.GinRequest) {
	form := &forms.CreateResourceAccountForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.CreateResourceAccount(c.Service(), form))
}

func (ResourceAccount) Delete(c *ctx.GinRequest) {
	form := &forms.DeleteResourceAccountForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.DeleteResourceAccount(c.Service(), form))
}

func (ResourceAccount) Update(c *ctx.GinRequest) {
	form := &forms.UpdateResourceAccountForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateResourceAccount(c.Service(), form))
}
