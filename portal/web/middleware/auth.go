package middleware

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models"
	"cloudiac/portal/services"
	"github.com/dgrijalva/jwt-go"
	"net/http"
)

// Auth 用户认证
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
		orgId := models.Id(c.GetHeader("IaC-Org-Id"))
		projectId := models.Id(c.GetHeader("IaC-Project-Id"))
		c.ServiceCtx().ProjectId = projectId
		c.ServiceCtx().OrgId = orgId
		c.ServiceCtx().UserId = claims.UserId
		c.ServiceCtx().Username = claims.Username
		c.ServiceCtx().IsSuperAdmin = claims.IsAdmin
		c.ServiceCtx().UserIpAddr = c.ClientIP()
		c.ServiceCtx().UserAgent = c.GetHeader("User-Agent")
	} else {
		c.JSONError(e.New(e.InvalidToken), http.StatusUnauthorized)
		return
	}
}

// AuthOrgId 验证组织ID是否有效
func AuthOrgId(c *ctx.GinRequestCtx) {
	if c.ServiceCtx().OrgId == "" {
		c.JSONError(e.New(e.InvalidOrganizationId), http.StatusForbidden)
		return
	}
	userOrgRel, err := services.FindUsersOrgRel(c.ServiceCtx().DB(), c.ServiceCtx().UserId, c.ServiceCtx().OrgId)
	if err == nil && len(userOrgRel) > 0 {
		c.ServiceCtx().Role = userOrgRel[0].Role
		c.Next()
		return
	}
	if c.ServiceCtx().IsSuperAdmin == true {
		c.Next()
		return
	}
	c.JSONError(e.New(e.PermissionDeny), http.StatusForbidden)
	return
}

func IsSuperAdmin(c *ctx.GinRequestCtx) {
	if c.ServiceCtx().IsSuperAdmin == true {
		c.Next()
	} else {
		c.JSONError(e.New(e.PermissionDeny), http.StatusForbidden)
	}
	return
}

func IsOrgAdmin(c *ctx.GinRequestCtx) {
	if c.ServiceCtx().Role == "owner" || c.ServiceCtx().IsSuperAdmin == true {
		c.Next()
	} else {
		c.JSONError(e.New(e.PermissionDeny), http.StatusForbidden)
	}
	return
}

// AuthProjectId 验证项目ID是否有效
func AuthProjectId(c *ctx.GinRequestCtx) {
	if c.ServiceCtx().ProjectId == "" {
		c.JSONError(e.New(e.InvalidProjectId), http.StatusForbidden)
		return
	}
	// TODO: 查找项目角色
	//userOrgRel, err := services.FindUsersOrgRel(c.ServiceCtx().DB(), c.ServiceCtx().UserId, c.ServiceCtx().OrgId)
	//if err == nil && len(userOrgRel) > 0 {
	//	c.ServiceCtx().Role = userOrgRel[0].Role
	//	c.Next()
	//	return
	//}
	//c.JSONError(e.New(e.PermissionDeny), http.StatusForbidden)
	return
}
