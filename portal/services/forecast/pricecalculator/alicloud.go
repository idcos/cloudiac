// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package pricecalculator

import (
	"cloudiac/configs"
	"cloudiac/portal/consts"
	"cloudiac/portal/models"
	"cloudiac/portal/services/forecast/schema"
	"cloudiac/utils"
	"cloudiac/utils/logs"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/alibabacloud-go/tea/tea"
)

type AliCloud struct {
	AccessKeyId     *string
	AccessKeySecret *string
}

func NewAliCloudBillService(vg *models.VariableGroup, f func(provider string, vars models.VarGroupVariables) map[string]string) (*AliCloud, error) {
	resAccount := f(vg.Provider, vg.Variables)
	if resAccount == nil {
		return nil, fmt.Errorf("provider: %s, resource account is null", vg.Provider)
	}

	if resAccount[consts.AlicloudAK] == "" || resAccount[consts.AlicloudSK] == "" {
		return nil, fmt.Errorf("provider: %s, resource account not exist", vg.Provider)
	}

	return &AliCloud{
		AccessKeyId:     tea.String(resAccount[consts.AlicloudAK]),
		AccessKeySecret: tea.String(resAccount[consts.AlicloudSK]),
	}, nil
}

func (a *AliCloud) GetResourcePrice(r *schema.Resource) (CloudCostPriceResp, error) {
	logger := logs.Get().WithField("func", "GetResourcePrice")
	resp := CloudCostPriceResp{}

	header := &http.Header{}
	header.Set("Content-Type", "application/json")

	costUrl := fmt.Sprintf("%s/price/search", configs.Get().CostServe)
	result := Result{
		Address:  r.Name,
		Provider: r.Provider,
		Region:   r.Region,
	}

	resources := make([]Resource, 0)
	for _, v := range r.RequestData {
		resources = append(resources, Resource{
			Type:      v.Type,
			Attribute: v.Attribute,
		})
	}
	result.Resource = resources

	param := CloudCostPriceRequest{
		[]Result{result},
	}

	logger.Infof("%+v", param)

	respData, err := utils.HttpService(costUrl, "POST", header, param, 5, 30)
	if err != nil {
		return resp, err
	}

	logger.Infof(fmt.Sprintf("%s", string(respData)))

	if err := json.Unmarshal(respData, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

func GetPriceFromResponse(resp CloudCostPriceResp) (float32, error) {
	var sum float32 = 0.0
	for _, detail := range resp.Result.Results {
		// 优惠价格最为实际的价格
		for _, attr := range detail.PriceAttr {
			if _, ok := attr["price"]; ok {
				sum += float32(utils.Str2float(attr["price"]))
			}

		}

	}

	return sum, nil
}
