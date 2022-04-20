// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package alicloud

import (
	"cloudiac/portal/consts"
	"cloudiac/portal/models"
	"cloudiac/portal/services/forecast/schema"
	"fmt"
	bssopenapi20171214 "github.com/alibabacloud-go/bssopenapi-20171214/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/client"
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

func (a *AliCloud) GetResourcePrice(r *schema.Resource) (*bssopenapi20171214.GetPayAsYouGoPriceResponse, error) {
	config := &openapi.Config{
		AccessKeyId: a.AccessKeyId,
		AccessKeySecret: a.AccessKeySecret,
		Endpoint:        tea.String("business.aliyuncs.com"),
	}

	client, _err := bssopenapi20171214.NewClient(config)
	if _err != nil {
		return nil, _err
	}

	request := make([]*bssopenapi20171214.GetPayAsYouGoPriceRequestModuleList, 0, len(r.RequestData))
	for _, v := range r.RequestData {
		request = append(request, &bssopenapi20171214.GetPayAsYouGoPriceRequestModuleList{
			ModuleCode: tea.String(v.Name),
			PriceType:  tea.String("Hour"),
			Config:     tea.String(v.Value),
		})
	}

	getPayAsYouGoPriceRequest := &bssopenapi20171214.GetPayAsYouGoPriceRequest{
		ProductCode:      tea.String(r.PriceCode),
		ModuleList:       request,
		SubscriptionType: tea.String("PayAsYouGo"),
		ProductType:      tea.String(r.PriceType),
	}

	return client.GetPayAsYouGoPrice(getPayAsYouGoPriceRequest)
}
