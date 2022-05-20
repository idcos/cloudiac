// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package middleware

import (
	"cloudiac/common"
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models"
	"cloudiac/portal/services"
	"cloudiac/portal/services/rbac"
	"cloudiac/utils/logs"
	"fmt"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
)

func parseArgs(lg logs.Logger, args ...string) (string, string, string) {
	var sub, obj, act string
	if len(args) >= 3 {
		sub = args[0] // 角色
		obj = args[1] // 对象
		act = args[2] // 操作
	} else if len(args) == 2 {
		obj = args[0]
		act = args[1]
	} else if len(args) == 1 {
		act = args[0]
	}
	if !(sub == "" && obj == "" && act == "") {
		lg.Tracef("policy overwrites %s, %s, %s", sub, obj, act)
	}

	return sub, obj, act
}

func parseRes(requestURI string) string {
	// 通过 RequestURI 解析资源名称
	res := ""
	// 请求 /api/v1/users/:userId，
	// 匹配第三段的  ^^^^^^ users
	regex := regexp.MustCompile("^/[^/]+/[^/]+/([^/?#]+)")
	match := regex.FindStringSubmatch(requestURI)
	if len(match) == 2 {
		res = match[1]
	} else {
		res = "other"
	}

	return res
}

func getOpFromMethod(method string) string {
	op := ""
	switch method {
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
	return op
}

func getCtxOrgRole(s *ctx.ServiceContext) string {
	// 组织角色
	role := ""
	switch {
	case s.UserId == "":
		role = consts.RoleAnonymous
	case s.IsSuperAdmin:
		role = consts.RoleRoot
	// 临时处理系统管理员权限
	case s.UserId == consts.SysUserId:
		role = consts.OrgRoleAdmin
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

	return role
}

func getCtxProjectRole(s *ctx.ServiceContext) string {
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
	case s.ProjectId == "":
		role, err := services.GetUserHighestProjectRole(s.DB(), s.OrgId, s.UserId)
		if err != nil {
			s.Logger().Errorf("get user highest project role error: %s", err)
		} else {
			proj = role
		}
	}

	return proj
}

func rewriteACParams(op, act, res, obj, role, sub string) (string, string, string) {
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

	return action, object, role
}

func changeToDemoRole(s *ctx.ServiceContext, role, proj string) (string, string) {
	if !s.IsSuperAdmin && s.OrgId != "" && s.OrgId == models.Id(common.DemoOrgId) {
		role = consts.RoleDemo
		proj = consts.RoleDemo
	}

	return role, proj
}

// AccessControl 角色访问权限控制
func AccessControl(args ...string) gin.HandlerFunc {
	logger := logs.Get().WithField("func", "AccessControl")
	var sub, obj, act = parseArgs(logger, args...)

	return func(g *gin.Context) {
		c := ctx.NewGinRequest(g)
		s := c.Service()

		// 通过 RequestURI 解析资源名称
		res := parseRes(c.Request.RequestURI)

		// 通过 HTTP method 解析资源动作,
		op := getOpFromMethod(c.Request.Method)

		// 通过 service ctx 获取组织角色,项目角色
		role := getCtxOrgRole(s)
		proj := getCtxProjectRole(s)

		// 参数重写
		action, object, role := rewriteACParams(op, act, res, obj, role, sub)

		// 访问演示组织资源的时候切换到演示模式角色
		role, proj = changeToDemoRole(s, role, proj)

		// 根据 角色 和 项目角色 判断资源访问许可
		allow, err := rbac.Enforce(role, proj, object, action)
		if err != nil {
			logger.Errorf("error enforce %s,%s %s:%s, err %s", role, proj, object, action, err)
			c.JSONError(e.New(e.InternalError), http.StatusInternalServerError)
		}
		logger.Tracef("enforce, orgRole=%s, projectRole=%s, object=%s, action=%s, allow=%v",
			role, proj, object, action, allow)
		if allow {
			c.Next()
		} else {
			c.JSONError(e.New(e.PermissionDeny, fmt.Errorf("%s,%s not allowed to %s %s", role, proj, action, object)), http.StatusForbidden)
		}
	}
}
