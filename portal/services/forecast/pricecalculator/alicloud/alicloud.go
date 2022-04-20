package alicloud

import (
	"cloudiac/portal/services/forecast/schema"
	bssopenapi20171214 "github.com/alibabacloud-go/bssopenapi-20171214/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/client"
	"github.com/alibabacloud-go/tea/tea"
)

type AliCloud struct {
	AccessKeyId     *string
	AccessKeySecret *string
}

func (a *AliCloud) GetResourcePrice(r *schema.Resource) (*bssopenapi20171214.GetPayAsYouGoPriceResponse, error) {
	config := &openapi.Config{
		// 您的AccessKey ID
		AccessKeyId: a.AccessKeyId,
		// 您的AccessKey Secret
		AccessKeySecret: a.AccessKeySecret,
	}
	request := make([]*bssopenapi20171214.GetPayAsYouGoPriceRequestModuleList, 0, len(r.RequestData))
	for _, v := range r.RequestData {
		request = append(request, &bssopenapi20171214.GetPayAsYouGoPriceRequestModuleList{
			ModuleCode: tea.String(v.Name),
			PriceType:  tea.String("Hour"),
			Config:     tea.String(v.Value),
		})
	}
	//fmt.Println(request,"qqqqq")

	// 访问的域名
	config.Endpoint = tea.String("business.aliyuncs.com")
	client := &bssopenapi20171214.Client{}
	client, _err := bssopenapi20171214.NewClient(config)
	if _err != nil {
		return nil, _err
	}

	getPayAsYouGoPriceRequest := &bssopenapi20171214.GetPayAsYouGoPriceRequest{
		ProductCode:      tea.String(r.PriceCode),
		ModuleList:       request,
		SubscriptionType: tea.String("PayAsYouGo"),
		ProductType:      tea.String(r.PriceType),
	}

	return client.GetPayAsYouGoPrice(getPayAsYouGoPriceRequest)
}
