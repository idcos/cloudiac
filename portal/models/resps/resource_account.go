// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package resps

import "cloudiac/portal/models"

type SearchResourceAccountResp struct {
	models.ResourceAccount
	CtServiceIds []string `json:"ctServiceIds"`
}
