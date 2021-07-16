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
		if c.ServiceCtx().IsSuperAdmin ||
			services.UserHasOrgRole(c.ServiceCtx().UserId, c.ServiceCtx().OrgId, "") {
			c.ServiceCtx().OrgId = orgId
		}
		c.JSONError(e.New(e.PermissionDeny), http.StatusForbidden)
		return
	}
	projectId := models.Id(c.GetHeader("IaC-Project-Id"))
	if projectId != "" {
		c.ServiceCtx().ProjectId = projectId
		if project, err := services.GetProjectsById(c.ServiceCtx().DB(), projectId); err != nil {
			c.JSONError(e.New(e.ProjectNotExists), http.StatusBadRequest)
			return
		} else if project.OrgId != c.ServiceCtx().OrgId {
			c.JSONError(e.New(e.PermissionDeny), http.StatusForbidden)
			return
		}
		if c.ServiceCtx().IsSuperAdmin ||
			services.UserHasOrgRole(c.ServiceCtx().UserId, c.ServiceCtx().OrgId, consts.OrgRoleAdmin) ||
			services.UserHasProjectRole(c.ServiceCtx().UserId, c.ServiceCtx().OrgId, c.ServiceCtx().ProjectId, "") {
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

// AuthProjectId 验证项目ID是否有效
func AuthProjectId(c *ctx.GinRequestCtx) {
	if c.ServiceCtx().ProjectId == "" {
		c.JSONError(e.New(e.InvalidProjectId), http.StatusForbidden)
		return
	}
	return
}
