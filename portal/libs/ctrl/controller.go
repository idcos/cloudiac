package ctrl

import (
	"reflect"

	"github.com/gin-gonic/gin"

	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
)

type Controller interface {
	Create(c *ctx.GinRequestCtx)
	Delete(c *ctx.GinRequestCtx)
	Update(c *ctx.GinRequestCtx)
	Search(c *ctx.GinRequestCtx)
	Detail(c *ctx.GinRequestCtx)
}

// 每次调用 handler 函数都应该克隆一个新的 controller
func cloneCtrl(ctrl Controller) Controller {
	typ := reflect.TypeOf(ctrl).Elem()
	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	return reflect.New(typ).Interface().(Controller)
}

func Register(group *gin.RouterGroup, ctrl Controller) {
	group.POST("/create", func(c *gin.Context) {
		cloneCtrl(ctrl).Create(ctx.NewRequestCtx(c))
	})
	group.DELETE("/delete", func(c *gin.Context) {
		cloneCtrl(ctrl).Delete(ctx.NewRequestCtx(c))
	})
	group.PUT("/update", func(c *gin.Context) {
		cloneCtrl(ctrl).Update(ctx.NewRequestCtx(c))
	})
	group.GET("/search", func(c *gin.Context) {
		cloneCtrl(ctrl).Search(ctx.NewRequestCtx(c))
	})
	group.GET("/detail", func(c *gin.Context) {
		cloneCtrl(ctrl).Detail(ctx.NewRequestCtx(c))
	})
}

// 包装我们自己接收 ctx.GinRequestCtx 的 handler，返回 gin.HandlerFunc
func GinRequestCtxWrap(handler func(*ctx.GinRequestCtx)) gin.HandlerFunc {
	return func(c *gin.Context) {
		handler(ctx.NewRequestCtx(c))
	}
}

type BaseController struct{}

func (b *BaseController) Create(c *ctx.GinRequestCtx) {
	c.JSONError(e.New(e.NotImplement))
}

func (b *BaseController) Delete(c *ctx.GinRequestCtx) {
	c.JSONError(e.New(e.NotImplement))
}

func (b *BaseController) Update(c *ctx.GinRequestCtx) {
	c.JSONError(e.New(e.NotImplement))
}

func (b *BaseController) Search(c *ctx.GinRequestCtx) {
	c.JSONError(e.New(e.NotImplement))
}

func (b *BaseController) Detail(c *ctx.GinRequestCtx) {
	c.JSONError(e.New(e.NotImplement))
}
