package pricecalculator

import (
	"cloudiac/portal/consts"
	"cloudiac/portal/models"
	"cloudiac/portal/services/forecast/pricecalculator/alicloud"
	"cloudiac/portal/services/forecast/schema"
	"cloudiac/utils"
	"cloudiac/utils/logs"
	"fmt"
	bssopenapi20171214 "github.com/alibabacloud-go/bssopenapi-20171214/client"
)

type PriceService interface {
	GetResourcePrice(r *schema.Resource) (*bssopenapi20171214.GetPayAsYouGoPriceResponse, error)
}

func NewPriceService(vg *models.VariableGroup,providerName string) (PriceService,error) {
	switch vg.Provider {
	case providerName:
		return alicloud.NewAliCloudBillService(vg,parseResourceAccount)
	default:
		logs.Get().Errorf("price service unsupported provider %s", vg.Provider)
		return nil, fmt.Errorf("price service  unsupported provider %s", vg.Provider)

	}
}


func parseResourceAccount(provider string, vars models.VarGroupVariables) map[string]string {
	resp := make(map[string]string)
	for _, v := range vars {
		if utils.ArrayIsExistsStr(consts.BillProviderResAccount[provider], v.Name) {
			if v.Sensitive {
				value, err := utils.DecryptSecretVarForce(v.Value)
				if err != nil {
					logs.Get().Errorf("get resource as sk error: %s", err)
					return nil
				}
				resp[v.Name] = value
				continue
			}
			resp[v.Name] = v.Value
		}
	}
	return resp
}
