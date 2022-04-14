// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package services

import "cloudiac/portal/consts"

var roleLevels = map[string]int{
	consts.RoleRoot:            1000,
	consts.OrgRoleAdmin:        220,
	consts.OrgRoleMember:       210,
	consts.ProjectRoleManager:  140,
	consts.ProjectRoleApprover: 130,
	consts.ProjectRoleOperator: 120,
	consts.ProjectRoleGuest:    110,
	consts.RoleDemo:            20,
	consts.RoleLogin:           10,
	consts.RoleAnonymous:       0,
}

func GetRoleLevel(role string) int {
	if level, ok := roleLevels[role]; ok {
		return level
	}
	return 0
}

// GetHighestRole 从 roles 中获取最高权限的角色
func GetHighestRole(roles ...string) string {
	var (
		highestRole  string
		highestLevel int
	)

	for _, role := range roles {
		level := GetRoleLevel(role)
		if level > highestLevel {
			highestRole = role
			highestLevel = level
		}
	}
	return highestRole
}
