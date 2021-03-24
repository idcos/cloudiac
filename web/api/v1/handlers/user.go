package handlers

import (
	"cloudiac/apps"
	"cloudiac/consts/e"
	"cloudiac/libs/ctrl"
	"cloudiac/libs/ctx"
	"cloudiac/models/forms"
)

type User struct {
	ctrl.BaseController
}

func (User) Create(c *ctx.GinRequestCtx) {
	form := forms.CreateUserForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.CreateUser(c.ServiceCtx(), &form))
}

func (User) Search(c *ctx.GinRequestCtx) {
	form := forms.SearchUserForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.SearchUser(c.ServiceCtx(), &form))
}

func (User) Update(c *ctx.GinRequestCtx) {
	form := forms.UpdateUserForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateUser(c.ServiceCtx(), &form))
}

func (User) Delete(c *ctx.GinRequestCtx) {
	//form := forms.DeleteUserForm{}
	//if err := c.Bind(&form); err != nil {
	//	return
	//}
	//c.JSONResult(apps.DeleteUser(c.ServiceCtx(), &form))
	c.JSONError(e.New(e.NotImplement))
}

func (User) Detail(c *ctx.GinRequestCtx) {
	form := forms.DetailUserForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.UserDetail(c.ServiceCtx(), &form))
}

func (User) RemoveUserForOrg(c *ctx.GinRequestCtx) {
	form := forms.DeleteUserForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.DeleteUserOrgMap(c.ServiceCtx(), &form))
}

func (User) UserPassReset(c *ctx.GinRequestCtx) {
	form := forms.DetailUserForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.UserPassReset(c.ServiceCtx(), &form))
}

func (User) Login(c *ctx.GinRequestCtx) {
	form := forms.LoginForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.Login(c.ServiceCtx(), &form))
}
