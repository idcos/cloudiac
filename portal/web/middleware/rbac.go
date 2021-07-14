package middleware

import (
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
		c := ctx.NewRequestCtx(g)

		// 通过 RequestURI 解析资源名称
		res := ""
		// 请求 /api/v1/users/:userId，
		// 匹配第三段的  ^^^^^^ users
		regex := regexp.MustCompile("^/[^/]+/[^/]+/([^/]+)")
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
		enforcer := c.ServiceCtx().Enforcer()
		err := enforcer.LoadPolicy()
		if err != nil {
			logger.Errorf("error load rbac policy, err %s", err)
			c.JSONError(e.New(e.DBError), http.StatusInternalServerError)
			return
		}

		// 获取用户组织角色
		c.Logger().Errorf("role %s[%s], para %s, res %s", c.ServiceCtx().Role, c.ServiceCtx().ProjectRole, c.Param("id"), res)
		if c.ServiceCtx().Role == "" && res == "orgs" && c.Param("id") != "" { // 通过 path 获取组织角色
			orgId := models.Id(c.Param("id"))
			userOrgRel, err := services.FindUsersOrgRel(c.ServiceCtx().DB(), c.ServiceCtx().UserId, orgId)
			if err == nil && len(userOrgRel) > 0 {
				c.ServiceCtx().Role = userOrgRel[0].Role
				c.ServiceCtx().OrgId = orgId
			}
			if c.ServiceCtx().IsSuperAdmin == true {
				c.ServiceCtx().Role = consts.OrgRoleRoot
			}
		}
		if c.ServiceCtx().ProjectRole == "" && res == "projects" && c.Param("id") != "" { // 通过 path 获取项目角色
			projectId := models.Id(c.Param("id"))
			role, err := services.GetProjectRoleByUser(c.ServiceCtx().DB(), projectId, c.ServiceCtx().UserId)
			if err == nil && role != "" {
				c.ServiceCtx().ProjectRole = role
				c.ServiceCtx().ProjectId = projectId
			}
			if c.ServiceCtx().IsSuperAdmin == true {
				c.ServiceCtx().ProjectRole = consts.ProjectRoleOwner
			}
		}

		role := c.ServiceCtx().Role
		proj := c.ServiceCtx().ProjectRole

		if c.ServiceCtx().IsSuperAdmin {
			role = "root" // 平台管理员
		} else if role == "" {
			if c.ServiceCtx().UserId != "" {
				role = "login" // 登陆用户，无 orgId 信息
			} else {
				role = "anonymous" // 未登陆用户
			}
		}

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
