package resps

import "cloudiac/portal/models"

type SearchResourceAccountResp struct {
	models.ResourceAccount
	CtServiceIds []string `json:"ctServiceIds"`
}
