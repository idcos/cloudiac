package middleware

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/services"
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
	exists, tokens := services.TokenExists(c.ServiceCtx().DB(), tokenStr)
	if !exists {
		c.JSONError(e.New(e.InvalidToken), http.StatusUnauthorized)
		return
	}
	if tokens != nil {
		c.ServiceCtx().UserId = tokens.CreatorId
	}
}
