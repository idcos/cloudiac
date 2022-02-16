// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package ctrl

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
)

type GinController struct{}

func (b *GinController) Create(c *ctx.GinRequest) {
	c.JSONError(e.New(e.NotImplement))
}

func (b *GinController) Delete(c *ctx.GinRequest) {
	c.JSONError(e.New(e.NotImplement))
}

func (b *GinController) Update(c *ctx.GinRequest) {
	c.JSONError(e.New(e.NotImplement))
}

func (b *GinController) Search(c *ctx.GinRequest) {
	c.JSONError(e.New(e.NotImplement))
}

func (b *GinController) Detail(c *ctx.GinRequest) {
	c.JSONError(e.New(e.NotImplement))
}
