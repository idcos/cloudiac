package middleware

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models"
	"cloudiac/portal/services"
	"cloudiac/utils/logs"
	"github.com/gin-gonic/gin"
	"net/http"
	"regexp"
)

func AC(args ...string) gin.HandlerFunc {
	logger := logs.Get().WithField("func", "initPolicy")

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
		logger.Debugf("policy %s %s %s", sub, obj, act)
	}

	return func(g *gin.Context) {
		c := ctx.NewRequestCtx(g)

		// 通过 RequestURI 解析资源名称
		if obj == "" {
			// 请求 /api/v1/users/:userId，
			// 匹配第三段的  ^^^^^^ users
			regex := regexp.MustCompile("^/[^/]+/[^/]+/([^/]+)")
			match := regex.FindStringSubmatch(c.Request.RequestURI)
			if len(match) == 2 {
				obj = match[1]
			} else {
				obj = "other"
			}
		}
		// 通过 HTTP method 解析资源动作
		if act == "" {
			switch c.Request.Method {
			case "GET":
				act = "read"
			case "POST":
				act = "create"
			case "PUT":
				act = "update"
			case "PATCH":
				act = "update"
			case "DELETE":
				act = "delete"
			default:
				act = "other"
			}
		}

		// 加载权限策略
		enforcer := c.ServiceCtx().Enforcer()
		err := enforcer.LoadPolicy()
		if err != nil {
			c.Logger().Errorf("error load rbac policy, err %s", err)
			c.JSONError(e.New(e.DBError), http.StatusInternalServerError)
			return
		}

		// 获取用户组织角色
		c.Logger().Debugf("id %s %s", c.Params, c.Param("id"))
		c.Logger().Debugf("userid %s orgid %s", c.ServiceCtx().UserId, c.ServiceCtx().OrgId)
		if obj == "orgs" && c.Param("id") != "" { // 通过 path 获取组织角色
			orgId := models.Id(c.Param("id"))
			userOrgRel, err := services.FindUsersOrgRel(c.ServiceCtx().DB(), c.ServiceCtx().UserId, orgId)
			c.Logger().Debugf("userOrgRel %s", userOrgRel)
			if err == nil && len(userOrgRel) > 0 {
				c.ServiceCtx().Role = userOrgRel[0].Role
				c.ServiceCtx().OrgId = orgId
			}
		} else if obj == "projects" && c.Param("id") != "" { // 通过 path 获取项目角色
			projectId := models.Id(c.Param("id"))
			// TODO: 获取项目角色
			//userOrgRel, err := services.FindUsersOrgRel(c.ServiceCtx().DB(), c.ServiceCtx().UserId, orgId)
			//if err == nil && len(userOrgRel) > 0 {
			//	c.ServiceCtx().ProjectRole = userOrgRel[0].Role
			c.ServiceCtx().ProjectId = projectId
			//}
		}

		role := c.ServiceCtx().Role
		proj := c.ServiceCtx().ProjectRole

		if c.ServiceCtx().IsSuperAdmin {
			role = "root" // 平台管理员
		}
		if role == "" {
			if c.ServiceCtx().OrgId == "" && c.ServiceCtx().UserId != "" {
				role = "login" // 登陆用户，无 orgId
			} else {
				role = "anonymous" // 未登陆用户
			}
		}

		// 根据 角色 和 项目角色 判断资源访问许可
		c.Logger().Infof("role %s proj %s, url %s method %s", role, proj, obj, act)
		allow, err := enforcer.Enforce(role, proj, obj, act)
		if err != nil {
			c.Logger().Errorf("error enforce %s", err)
			c.JSONError(e.New(e.InternalError), http.StatusInternalServerError)
		}
		if allow {
			c.Next()
		} else {
			c.JSONError(e.New(e.PermissionDeny), http.StatusForbidden)
		}
	}
}
