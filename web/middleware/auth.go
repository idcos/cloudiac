package middleware

import (
	"cloudiac/consts/e"
	"cloudiac/libs/ctx"
	"cloudiac/services"
	"net/http"
	"strconv"

	"github.com/dgrijalva/jwt-go"
)

// 用户认证
func Auth(c *ctx.GinRequestCtx) {
	tokenStr := c.GetHeader("Authorization")
	if tokenStr == "" {
		c.Logger().Infof("missing token")
		c.JSONError(e.New(e.InvalidToken), http.StatusUnauthorized)
		return
	}

	token, err := jwt.ParseWithClaims(tokenStr, &services.Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(services.SecretKey), nil
	})
	if err != nil {
		c.JSONError(e.New(e.InvalidToken), http.StatusUnauthorized)
		return
	}

	if claims, ok := token.Claims.(*services.Claims); ok && token.Valid {
		orgId, _ := strconv.ParseUint(c.GetHeader("IaC-Org-Id"), 10, 32)
		c.ServiceCtx().OrgId = uint(orgId)
		c.ServiceCtx().UserId = claims.UserId
		c.ServiceCtx().Username = claims.Username
		c.ServiceCtx().IsAdmin = claims.IsAdmin
		c.ServiceCtx().UserIpAddr = c.ClientIP()
		c.ServiceCtx().UserAgent = c.GetHeader("User-Agent")
	} else {
		c.JSONError(e.New(e.InvalidToken), http.StatusUnauthorized)
		return
	}
}

// 验证组织ID是否有效
func AuthOrgId(c *ctx.GinRequestCtx) {
	if c.ServiceCtx().OrgId == 0 {
		c.JSONError(e.New(e.InvalidOrganizationId), http.StatusForbidden)
		return
	}
	userOrgMap, err := services.FindUsersOrgMap(c.ServiceCtx().DB(), c.ServiceCtx().UserId, c.ServiceCtx().OrgId)
	if err != nil || len(userOrgMap) == 0 {
		c.JSONError(e.New(e.PermissionDeny), http.StatusForbidden)
		return
	}
	c.ServiceCtx().Role = userOrgMap[0].Role
	c.Next()
	return
}

func IsAdmin(c *ctx.GinRequestCtx) {
	if c.ServiceCtx().IsAdmin == true {
		c.Next()
	} else {
		c.JSONError(e.New(e.PermissionDeny), http.StatusForbidden)
	}
	return
}

func IsOrgOwner(c *ctx.GinRequestCtx)  {
	if c.ServiceCtx().Role == "owner" {
		c.Next()
	} else {
		c.JSONError(e.New(e.PermissionDeny), http.StatusForbidden)
	}
	return
}