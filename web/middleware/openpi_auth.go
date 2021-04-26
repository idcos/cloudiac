package middleware

import (
	"cloudiac/consts/e"
	"cloudiac/libs/ctx"
	"cloudiac/services"
	"net/http"
)

//apitoken认证
func OpenApiAuth(c *ctx.GinRequestCtx) {
	tokenStr := c.GetHeader("Authorization")
	if tokenStr == "" {
		c.Logger().Infof("missing token")
		c.JSONError(e.New(e.InvalidToken), http.StatusUnauthorized)
		return
	}
	//校验apitoken
	if services.TokenExists(c.ServiceCtx().DB(),tokenStr){
		c.JSONError(e.New(e.InvalidToken), http.StatusUnauthorized)
		return
	}
}
