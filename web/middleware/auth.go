// +build !noauth

package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"cloudiac/consts/e"
	"cloudiac/libs/ctx"
	"cloudiac/services"
)

var perms = map[string][]string{
	"ListInstances": {
		"/metric/search",
		"/metric/type/search",
		"/metric/history/*",
		"/metrics/search",
	},
	"ListHostDashboard": []string{
		"/metric/search",
		"/metric/type/search",
		"/metric/history/*",
		"/metric/topk/*",
		"/metrics/search",
		"/datasource/search",
	},
}

type Data struct {
	Token string `json:"token"`
}

func matchPerm(apiPath string, userPerms []string) bool {
	for _, perm := range userPerms {
		if perm == "admin" {
			return true
		}

		for _, api := range perms[perm] {
			if api == apiPath || (strings.HasSuffix(api, "*") &&
				strings.HasPrefix(apiPath, strings.TrimSuffix(api, "*"))) {
				return true
			}
		}
	}
	return false
}

// 用户认证
func Auth(c *ctx.GinRequestCtx) {
	if os.Getenv("IMS_NO_AUTH") == "1" {
		sc := c.ServiceCtx()
		sc.Token = ""
		sc.TenantId = 1
		sc.UserId = 0
		sc.Username = "admin"
		sc.UserIpAddr = c.ClientIP()
		sc.UserAgent = c.GetHeader("User-Agent")
		sc.AddLogField("uid", fmt.Sprintf("%d", sc.UserId))
		for k := range perms {
			c.ServiceCtx().Perms = append(c.ServiceCtx().Perms, k)
		}
		c.Next()
		return
	}

	token := c.GetHeader("Authorization")
	if token == "" {
		c.Logger().Infof("missing token")
		c.JSONError(e.New(e.InvalidToken), http.StatusUnauthorized)
		return
	}
	userInfo, er := services.GetUserInfo(token)
	if er != nil {
		c.Logger().Warnf("token verify failed: %v", er)
		c.JSONError(e.New(e.ValidateError), http.StatusUnauthorized)
		return
	}
	// 验证组织ID是否合法
	tenantId, _ := strconv.ParseUint(c.GetHeader("Ops-Tenant-Id"), 10, 32)
	if userInfo.TenantId != uint(tenantId) {
		c.Logger().Infof("invalid tenantId")
		c.JSONError(e.New(e.InvalidTenantId), http.StatusForbidden)
		return
	}

	c.Logger().Tracef("userInfo.Permissions: %#v", userInfo.Permissions)
	c.ServiceCtx().Perms = ParsePerms(userInfo.Permissions)
	c.ServiceCtx().Token = token
	c.ServiceCtx().TenantId = uint(tenantId)
	c.ServiceCtx().UserId = userInfo.Id
	c.ServiceCtx().Username = userInfo.Username
	c.ServiceCtx().UserIpAddr = c.ClientIP()
	c.ServiceCtx().UserAgent = c.GetHeader("User-Agent")
	c.ServiceCtx().AddLogField("uid", fmt.Sprintf("%d", userInfo.Id))
}

func ParsePerms(permissions []services.Permission) []string {
	perms := make([]string, 0)
	for _, v := range permissions {
		perms = append(perms, v.Actions...)
	}
	return perms
}

// api 鉴权
func ApiAuth(c *ctx.GinRequestCtx) {
	apiPath := strings.TrimPrefix(c.Request.URL.Path, "/api/v1")
	if matchPerm(apiPath, c.ServiceCtx().Perms) {
		c.Next()
		return
	}
	c.JSONError(e.New(e.PermissionDeny), http.StatusForbidden)
}
