// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package pricecalculator

import (
	"cloudiac/portal/consts"
	"cloudiac/portal/models"
	"cloudiac/portal/services/forecast/schema"
	"cloudiac/utils"
	"cloudiac/utils/logs"
	"fmt"
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
