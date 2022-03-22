package resps

import "cloudiac/portal/models"

type RespTplOfPolicy struct {
	models.Policy
	GroupName string `json:"groupName"`
	GroupId   string `json:"groupId"`
	TplName   string `json:"tplName"`
}

type RespTplOfPolicyGroup struct {
	GroupName string `json:"groupName"`
	GroupId   string `json:"groupId"`
}
