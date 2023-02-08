// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package resps

import "cloudiac/portal/models"

type UserWithRoleResp struct {
	models.User
	Password    string `json:"-"`
	Role        string `json:"role,omitempty" example:"member"`         // 组织角色
	ProjectRole string `json:"projectRole,omitempty" example:"manager"` // 项目角色
}

type CreateUserResp struct {
	*models.User
	InitPass string `json:"initPass,omitempty" example:"rANd0m"` // 初始化密码
}
