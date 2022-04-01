package billcollect

import (
	"cloudiac/portal/models"
	bssopenapi20171214 "github.com/alibabacloud-go/bssopenapi-20171214/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/pkg/errors"
)

func newAliBillInstance(vg *models.VariableGroup) (*aliClint, error) {
	config := &openapi.Config{
		AccessKeyId:     &ab.Ak,
		AccessKeySecret: &ab.Sk,
		Endpoint:        tea.String("business.aliyuncs.com"),
	}

	result, err := bssopenapi20171214.NewClient(config)
	if err != nil {
		return nil, err
	}
	return &aliClint{
		clint:    result,
		pageNum:  1,
		pageSize: 300,
	}, err
}

type aliBill struct {
	Ak string
	Sk string
}

func (ab *aliBill) Clint() (ClintIface, error) {

}

type aliClint struct {
	clint    *bssopenapi20171214.Client
	pageNum  int32
	pageSize int32
}

func (ac *aliClint) GetResourceMonthCost(billingCycle string) ([]ResourceCost, error) {
	queryInstanceBillRequest := &bssopenapi20171214.QueryInstanceBillRequest{
		//BillingCycle: tea.String("2022-03"),
		BillingCycle: tea.String(billingCycle),
		PageNum:      tea.Int32(ac.pageNum),
		//单次请求最大值300
		PageSize: tea.Int32(ac.pageSize),
	}

	result, err := ac.clint.QueryInstanceBill(queryInstanceBillRequest)
	if err != nil {
		return nil, err
	}
	if !*result.Body.Success {
		return nil, errors.New("")
	}

	resp := make([]ResourceCost, 0)
	for _, v := range result.Body.Data.Items.Item {
		resp = append(resp, ResourceCost{
			ProductCode:    *v.ProductCode,
			InstanceId:     *v.InstanceID,
			InstanceConfig: *v.InstanceConfig,
			PretaxAmount:   *v.PretaxAmount,
			Region:         *v.Region,
			Currency:       *v.Currency,
			Cycle:          billingCycle,
			Provider:       "alicloud",
		})
	}

	if *result.Body.Data.TotalCount > (ac.pageSize * ac.pageNum) {
		ac.pageNum++
		r, _ := ac.GetResourceMonthCost(billingCycle)
		resp = append(resp, r...)
	}

	return nil, err
}

func (ac *aliClint) GetResourceDayCost(billingCycle string) ([]ResourceCost, error) {
	return nil, nil
}
