package pricecalculator

import (
	"cloudiac/portal/services/forecast/pricecalculator/alicloud"
	"cloudiac/portal/services/forecast/schema"
	bssopenapi20171214 "github.com/alibabacloud-go/bssopenapi-20171214/client"
	"github.com/alibabacloud-go/tea/tea"
)

type PriceService interface {
	GetResourcePrice(r *schema.Resource) (*bssopenapi20171214.GetPayAsYouGoPriceResponse, error)
}

func NewPriceService(providerName string) PriceService {
	//todo 根据provider匹配对应的计算接口
	return &alicloud.AliCloud{
		AccessKeyId:     tea.String(""),
		AccessKeySecret: tea.String(""),
	}
}
