// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package resps

import "cloudiac/portal/models"

type KeyResp struct {
	models.Key
	Creator string `json:"creator"`
}
