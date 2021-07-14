package middleware

import (
	"cloudiac/portal/consts"
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
		c.ServiceCtx().UserId = claims.UserId
		c.ServiceCtx().Username = claims.Username
		c.ServiceCtx().IsSuperAdmin = claims.IsAdmin
		c.ServiceCtx().UserIpAddr = c.ClientIP()
		c.ServiceCtx().UserAgent = c.GetHeader("User-Agent")
	} else {
		c.JSONError(e.New(e.InvalidToken), http.StatusUnauthorized)
		return
	}

	orgId := models.Id(c.GetHeader("IaC-Org-Id"))
	if orgId != "" {
		c.ServiceCtx().OrgId = orgId
		userOrgRel, err := services.FindUsersOrgRel(c.ServiceCtx().DB(), c.ServiceCtx().UserId, c.ServiceCtx().OrgId)
		if err == nil && len(userOrgRel) > 0 {
			c.ServiceCtx().Role = userOrgRel[0].Role
			c.Next()
			return
		}
		if c.ServiceCtx().IsSuperAdmin == true {
			c.ServiceCtx().Role = consts.OrgRoleRoot
			c.Next()
			return
		}
		c.JSONError(e.New(e.PermissionDeny), http.StatusForbidden)
		return
	}
	projectId := models.Id(c.GetHeader("IaC-Project-Id"))
	if projectId != "" {
		c.ServiceCtx().ProjectId = projectId
		role, err := services.GetProjectRoleByUser(c.ServiceCtx().DB(), c.ServiceCtx().ProjectId, c.ServiceCtx().UserId)
		if err == nil && role != "" {
			c.ServiceCtx().ProjectRole = role
			c.Next()
			return
		}
		if c.ServiceCtx().IsSuperAdmin == true {
			c.ServiceCtx().ProjectRole = consts.ProjectRoleOwner
			c.Next()
			return
		}
		c.JSONError(e.New(e.PermissionDeny), http.StatusForbidden)
		return
	}
}

// AuthOrgId 验证组织ID是否有效
func AuthOrgId(c *ctx.GinRequestCtx) {
	if c.ServiceCtx().OrgId == "" {
		c.JSONError(e.New(e.InvalidOrganizationId), http.StatusForbidden)
		return
	}
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
	if c.ServiceCtx().Role == consts.OrgRoleAdmin || c.ServiceCtx().IsSuperAdmin == true {
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
	return
}
