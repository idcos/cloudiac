package v1

// 用户管理示例

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
	//form := &forms.CreateUserForm{}
	//if err := c.Bind(form); err != nil {
	//	return
	//}
	//c.JSONResult(apps.CreateUser(c.ServiceCtx(), form))
	c.JSONError(e.New(e.NotImplement))
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
