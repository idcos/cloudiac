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
)

type CloudCostPriceRequest struct {
	Results []Result `json:"results"`
}

type Result struct {
	Address  string     `json:"address"`
	Provider string     `json:"provider"`
	Region   string     `json:"region"`
	Resource []Resource `json:"resource"`
}

type Resource struct {
	Type      string            `json:"type"`
	Attribute map[string]string `json:"attribute"`
}

type PriceResp struct {
	Results []ResultResp `json:"results"`
}

type ResultResp struct {
	Address   string              `json:"address"`
	PriceAttr []map[string]string `json:"priceAttr"`
}

type CloudCostPriceResp struct {
	Code    int       `json:"code"`
	Message string    `json:"message"`
	Result  PriceResp `json:"result"`
}

type PriceService interface {
	GetResourcePrice(r *schema.Resource) (CloudCostPriceResp, error)
}

func NewPriceService(vg *models.VariableGroup, providerName string) (PriceService, error) {
	switch vg.Provider {
	case providerName:
		return NewAliCloudBillService(vg, parseResourceAccount)
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

func GetResourcePrice(r *schema.Resource) (CloudCostPriceResp, error) {
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
		for _, attr := range detail.PriceAttr {
			if _, ok := attr["price"]; ok {
				sum += float32(utils.Str2float(attr["price"]))
			}

		}

	}

	return sum, nil
}
