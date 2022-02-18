// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package middleware

import (
	"cloudiac/configs"
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models"
	"cloudiac/portal/services"
	"fmt"
	"net/http"

	"github.com/dgrijalva/jwt-go"
)

func checkToken(c *ctx.GinRequest, tokenStr string) (models.Id, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &services.Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(configs.Get().JwtSecretKey), nil
	})

	var apiTokenOrgId models.Id
	if err != nil || token == nil {
		apiToken, err := services.GetApiTokenByToken(c.Service().DB(), tokenStr)
		if err != nil {
			return apiTokenOrgId, err
		}
		c.Service().UserId = consts.SysUserId
		c.Service().Username = consts.DefaultSysName
		c.Service().IsSuperAdmin = false
		c.Service().UserIpAddr = c.ClientIP()
		apiTokenOrgId = apiToken.OrgId
		return apiTokenOrgId, nil
	}

	if claims, ok := token.Claims.(*services.Claims); ok && token.Valid &&
		claims.Subject == consts.JwtSubjectUserAuth {

		c.Service().UserId = claims.UserId
		c.Service().Username = claims.Username
		c.Service().IsSuperAdmin = claims.IsAdmin
		c.Service().UserIpAddr = c.ClientIP()
	} else {
		return apiTokenOrgId, e.New(e.InvalidToken)
	}
	return apiTokenOrgId, nil
}

func checkOrgId(c *ctx.GinRequest, orgId, apiTokenOrgId models.Id) (e.Error, int) {
	if orgId == "" {
		return nil, -1
	}

	c.Service().OrgId = orgId
	// 校验api token所属组织是否与传入组织一致
	if apiTokenOrgId != "" && !(orgId == apiTokenOrgId) {
		return e.New(e.InvalidToken), http.StatusUnauthorized
	}

	if org, err := services.GetOrganizationById(c.Service().DB(), orgId); err != nil {
		return e.New(e.OrganizationNotExists, fmt.Errorf("not allow to access org")), http.StatusBadRequest
	} else if org.Status == models.Disable && !c.Service().IsSuperAdmin {
		return e.New(e.PermissionDeny, fmt.Errorf("org disabled")), http.StatusForbidden
	}
	if c.Service().IsSuperAdmin ||
		services.UserHasOrgRole(c.Service().UserId, c.Service().OrgId, "") {
	} else {
		return e.New(e.PermissionDeny, fmt.Errorf("not allow to access org")), http.StatusForbidden
	}
	return nil, -1
}

// Auth 用户认证
func Auth(c *ctx.GinRequest) {
	tokenStr := c.GetHeader("Authorization")
	if tokenStr == "" {
		c.Logger().Infof("missing token")
		c.JSONError(e.New(e.InvalidToken), http.StatusUnauthorized)
		return
	}

	apiTokenOrgId, err := checkToken(c, tokenStr)
	if err != nil {
		c.JSONError(e.New(e.InvalidToken), http.StatusUnauthorized)
		return
	}

	orgId := models.Id(c.GetHeader("IaC-Org-Id"))
	if err, httpCode := checkOrgId(c, orgId, apiTokenOrgId); err != nil {
		c.JSONError(err, httpCode)
		return
	}

	projectId := models.Id(c.GetHeader("IaC-Project-Id"))
	if projectId == "" {
		return
	}

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

// AuthOrgId 验证组织ID是否有效
func AuthOrgId(c *ctx.GinRequest) {
	if c.Service().OrgId == "" {
		c.JSONError(e.New(e.InvalidOrganizationId), http.StatusForbidden)
		return
	}
}

// AuthProjectId 验证项目ID是否有效
func AuthProjectId(c *ctx.GinRequest) {
	if c.Service().ProjectId == "" {
		c.JSONError(e.New(e.InvalidProjectId), http.StatusForbidden)
		return
	}
}

func AuthApiToken(c *ctx.GinRequest) {
	token, _ := c.GetQuery("token")
	if _, err := services.IsActiveToken(c.Service().DB(), token, consts.TokenTrigger); err != nil {
		c.JSONError(e.New(e.PermissionDeny, fmt.Errorf("missing token")), http.StatusForbidden)
		return
	}
}
