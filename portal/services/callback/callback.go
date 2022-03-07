package callback

import "cloudiac/portal/models"

type ResourceResult struct {
	Resources []models.Resource `json:"resources"  `
}

type IacCallbackContent struct {
	ExtraData  models.JSON    `json:"extraData"`
	TaskStatus string         `json:"taskStatus"`
	OrgId      models.Id      `json:"orgId"`
	ProjectId  models.Id      `json:"projectId"`
	TplId      models.Id      `json:"tplId"`
	EnvId      models.Id      `json:"envId"`
	Result     ResourceResult `json:"result"`
}

func SendCallback(){

}