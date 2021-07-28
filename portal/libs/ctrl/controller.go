package ctrl

import (
	"reflect"

	"github.com/gin-gonic/gin"

	"cloudiac/portal/libs/ctx"
)

type Controller interface {
	Create(c *ctx.GinRequest)
	Delete(c *ctx.GinRequest)
	Update(c *ctx.GinRequest)
	Search(c *ctx.GinRequest)
	Detail(c *ctx.GinRequest)
}

func newController(ctrl Controller) Controller {
	typ := reflect.TypeOf(ctrl).Elem()
	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	return reflect.New(typ).Interface().(Controller)
}

func WrapHandler(handler func(*ctx.GinRequest)) gin.HandlerFunc {
	return func(c *gin.Context) {
		handler(ctx.NewGinRequest(c))
	}
}

func Register(group *gin.RouterGroup, ctrl Controller) {
	group.POST("", func(c *gin.Context) {
		newController(ctrl).Create(ctx.NewGinRequest(c))
	})
	group.DELETE("/:id", func(c *gin.Context) {
		newController(ctrl).Delete(ctx.NewGinRequest(c))
	})
	group.PUT("/:id", func(c *gin.Context) {
		newController(ctrl).Update(ctx.NewGinRequest(c))
	})
	group.GET("/:id", func(c *gin.Context) {
		newController(ctrl).Detail(ctx.NewGinRequest(c))
	})
	group.GET("", func(c *gin.Context) {
		newController(ctrl).Search(ctx.NewGinRequest(c))
	})
}
