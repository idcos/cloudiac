// Copyright 2021 CloudJ Company Limited. All rights reserved.

package middleware

import (
	"cloudiac/common"
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models"
	"cloudiac/portal/services"
	"cloudiac/utils/logs"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"regexp"
)

// AccessControl 角色访问权限控制
func AccessControl(args ...string) gin.HandlerFunc {
	logger := logs.Get().WithField("func", "AccessControl")

	var sub, obj, act string
	if len(args) >= 3 {
		sub = args[0]
		obj = args[1]
		act = args[2]
	} else if len(args) == 2 {
		obj = args[0]
		act = args[1]
	} else if len(args) == 1 {
		act = args[0]
	}
	if !(sub == "" && obj == "" && act == "") {
		logger.Debugf("policy overwrites %s, %s, %s", sub, obj, act)
	}

	return func(g *gin.Context) {
		c := ctx.NewGinRequest(g)
		s := c.Service()

		// 通过 RequestURI 解析资源名称
		res := ""
		// 请求 /api/v1/users/:userId，
		// 匹配第三段的  ^^^^^^ users
		regex := regexp.MustCompile("^/[^/]+/[^/]+/([^/?#]+)")
		match := regex.FindStringSubmatch(c.Request.RequestURI)
		if len(match) == 2 {
			res = match[1]
		} else {
			res = "other"
		}

		// 通过 HTTP method 解析资源动作
		op := "read"
		switch c.Request.Method {
		case "GET":
			op = "read"
		case "POST":
			op = "create"
		case "PUT":
			op = "update"
		case "PATCH":
			op = "update"
		case "DELETE":
			op = "delete"
		default:
			op = "other"
		}

		// 加载权限策略
		enforcer := c.Service().Enforcer()
		err := enforcer.LoadPolicy()
		if err != nil {
			logger.Errorf("error load rbac policy, err %s", err)
			c.JSONError(e.New(e.DBError), http.StatusInternalServerError)
			return
		}

		// 组织角色
		role := ""
		switch {
		case s.UserId == "":
			role = consts.RoleAnonymous
		case s.IsSuperAdmin:
			role = consts.RoleRoot
		case s.UserId != "" && s.OrgId == "":
			role = consts.RoleLogin
		case s.OrgId != "":
			userOrgs := services.UserOrgRoles(s.UserId)
			userOrg := userOrgs[s.OrgId]
			if userOrg != nil {
				role = userOrg.Role
			}
		default:
		}
		//s.Role = role

		// 项目角色
		proj := ""
		switch {
		case s.IsSuperAdmin:
			proj = consts.ProjectRoleManager
		case services.UserHasOrgRole(s.UserId, s.OrgId, consts.OrgRoleAdmin):
			proj = consts.ProjectRoleManager
		case s.ProjectId != "":
			userProjects := services.UserProjectRoles(s.UserId)
			userProject := userProjects[s.ProjectId]
			if userProject != nil {
				proj = userProject.Role
			}
		default:
		}
		//s.ProjectRole = proj

		// 参数重写
		action := op
		if act != "" {
			action = act
		}
		object := res
		if obj != "" {
			object = obj
		}
		if sub != "" {
			role = sub
		}

		// 访问演示组织资源的时候切换到演示模式角色
		if !s.IsSuperAdmin && s.OrgId != "" && s.OrgId == models.Id(common.DemoOrgId) {
			role = consts.RoleDemo
			proj = consts.RoleDemo
		}

		// 根据 角色 和 项目角色 判断资源访问许可
		logger.Debugf("enforcing %s,%s %s:%s", role, proj, object, action)
		allow, err := enforcer.Enforce(role, proj, object, action)
		if err != nil {
			logger.Errorf("error enforce %s,%s %s:%s, err %s", role, proj, object, action, err)
			c.JSONError(e.New(e.InternalError), http.StatusInternalServerError)
		}
		if allow {
			c.Next()
		} else {
			c.JSONError(e.New(e.PermissionDeny, fmt.Errorf("%s,%s not allowed to %s %s", role, proj, action, object)), http.StatusForbidden)
		}
	}
}
