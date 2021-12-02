// Copyright 2021 CloudJ Company Limited. All rights reserved.

package middleware

import (
	"cloudiac/configs"
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models"
	"cloudiac/portal/services"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"net/http"
)

// Auth 用户认证
func Auth(c *ctx.GinRequest) {
	tokenStr := c.GetHeader("Authorization")
	var apiTokenOrgId models.Id
	if tokenStr == "" {
		c.Logger().Infof("missing token")
		c.JSONError(e.New(e.InvalidToken), http.StatusUnauthorized)
		return
	}

	err := func() error {
		token, err := jwt.ParseWithClaims(tokenStr, &services.Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(configs.Get().JwtSecretKey), nil
		})
		if err != nil || token == nil {
			apiToken, err := services.GetApiTokenByToken(c.Service().DB(), tokenStr)
			if err != nil {
				return err
			}
			c.Service().UserId = consts.SysUserId
			c.Service().Username = consts.DefaultSysName
			c.Service().IsSuperAdmin = false
			c.Service().UserIpAddr = c.ClientIP()
			apiTokenOrgId = apiToken.OrgId
			return nil
		}

		if claims, ok := token.Claims.(*services.Claims); ok && token.Valid {
			c.Service().UserId = claims.UserId
			c.Service().Username = claims.Username
			c.Service().IsSuperAdmin = claims.IsAdmin
			c.Service().UserIpAddr = c.ClientIP()
		} else {
			c.JSONError(e.New(e.InvalidToken), http.StatusUnauthorized)
		}
		return nil
	}()

	if err != nil {
		c.JSONError(e.New(e.InvalidToken), http.StatusUnauthorized)
		return
	}
	orgId := models.Id(c.GetHeader("IaC-Org-Id"))
	if orgId != "" {
		c.Service().OrgId = orgId
		// 校验api token所属组织是否与传入组织一致
		if apiTokenOrgId != "" && !(orgId == apiTokenOrgId) {
			c.JSONError(e.New(e.InvalidToken), http.StatusUnauthorized)
			return
		}

		if org, err := services.GetOrganizationById(c.Service().DB(), orgId); err != nil {
			c.JSONError(e.New(e.OrganizationNotExists, fmt.Errorf("not allow to access org")), http.StatusBadRequest)
			return
		} else if org.Status == models.Disable && !c.Service().IsSuperAdmin {
			c.JSONError(e.New(e.PermissionDeny, fmt.Errorf("org disabled")), http.StatusForbidden)
			return
		}
		if c.Service().IsSuperAdmin ||
			services.UserHasOrgRole(c.Service().UserId, c.Service().OrgId, "") {
		} else {
			c.JSONError(e.New(e.PermissionDeny, fmt.Errorf("not allow to access org")), http.StatusForbidden)
			return
		}
	}
	projectId := models.Id(c.GetHeader("IaC-Project-Id"))
	if projectId != "" {
		c.Service().ProjectId = projectId
		if project, err := services.GetProjectsById(c.Service().DB(), projectId); err != nil {
			c.JSONError(e.New(e.ProjectNotExists, fmt.Errorf("not allow to access project")), http.StatusBadRequest)
			return
		} else if project.OrgId != c.Service().OrgId {
			c.JSONError(e.New(e.PermissionDeny, fmt.Errorf("invalid project id")), http.StatusForbidden)
			return
		} else if project.Status == models.Disable &&
			!(c.Service().IsSuperAdmin || services.UserHasOrgRole(c.Service().UserId, c.Service().OrgId, consts.OrgRoleAdmin)) {
			c.JSONError(e.New(e.PermissionDeny, fmt.Errorf("project disabled")), http.StatusForbidden)
			return
		}
		if c.Service().IsSuperAdmin ||
			services.UserHasOrgRole(c.Service().UserId, c.Service().OrgId, consts.OrgRoleAdmin) ||
			services.UserHasProjectRole(c.Service().UserId, c.Service().OrgId, c.Service().ProjectId, "") {
			c.Next()
			return
		}
		c.JSONError(e.New(e.PermissionDeny, fmt.Errorf("not allow to access project")), http.StatusForbidden)
		return
	}
}

// AuthOrgId 验证组织ID是否有效
func AuthOrgId(c *ctx.GinRequest) {
	if c.Service().OrgId == "" {
		c.JSONError(e.New(e.InvalidOrganizationId), http.StatusForbidden)
		return
	}
	return
}

// AuthProjectId 验证项目ID是否有效
func AuthProjectId(c *ctx.GinRequest) {
	if c.Service().ProjectId == "" {
		c.JSONError(e.New(e.InvalidProjectId), http.StatusForbidden)
		return
	}
	return
}

func AuthApiToken(c *ctx.GinRequest) {
	token, _ := c.GetQuery("token")
	if _, err := services.IsExistsTriggerToken(c.Service().DB(), token); err != nil {
		c.JSONError(e.New(e.PermissionDeny, fmt.Errorf("missing token")), http.StatusForbidden)
		return
	}
}
