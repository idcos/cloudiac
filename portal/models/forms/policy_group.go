package forms

import "cloudiac/portal/models"

type SearchRegistryPolicyGroupForm struct {
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"pageSize" form:"pageSize"`
	Q        string `json:"q" form:"q"`
}

type SearchRegistryPolicyGroupVersionForm struct {
	Namespace string `json:"ns" form:"ns" binding:"required"`
	GroupName string `json:"gn" form:"gn" binding:"required"`
}

type RegistryPolicyGroupVersionsResp struct {
	Id        models.Id `json:"id"`
	Namespace string    `json:"namespace"`
	GroupName string    `json:"groupName"`

	//  eg. 'v0.1.1' or '0.1.1'
	GitTag   string `json:"gitTag"`
	CommitId string `json:"commitId"`
}
