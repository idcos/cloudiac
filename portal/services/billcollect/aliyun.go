package billcollect

import (
	"cloudiac/portal/consts"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"fmt"
	bssopenapi20171214 "github.com/alibabacloud-go/bssopenapi-20171214/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/fatih/structs"
	"github.com/pkg/errors"
)

func NewAlicloudBillProvider(vg *models.VariableGroup) (*aliClint, error) {
	resAccount := parseResourceAccount(vg.Provider, vg.Variables)
	if resAccount == nil {
		return nil, fmt.Errorf("provider: %s, resource account is null", vg.Provider)
	}

	if resAccount[consts.AlicloudAK] == "" || resAccount[consts.AlicloudSK] == "" {
		return nil, fmt.Errorf("provider: %s, resource account not exist", vg.Provider)
	}

	config := &openapi.Config{
		AccessKeyId:     tea.String(resAccount[consts.AlicloudAK]),
		AccessKeySecret: tea.String(resAccount[consts.AlicloudSK]),
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
		provider: vg.Provider,
		vg:       vg,
	}, err
}

type aliBill struct {
	Ak string
	Sk string
}

type aliClint struct {
	clint    *bssopenapi20171214.Client
	provider string
	pageNum  int32
	pageSize int32
	vg       *models.VariableGroup
}

func (ac *aliClint) Provider() string {
	return ac.provider
}

func (ac *aliClint) GetResourceMonthCost(billingCycle string) ([]*bssopenapi20171214.QueryInstanceBillResponseBodyDataItemsItem, error) {
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
		return nil, errors.New("get %s bill error")
	}
	resp := make([]*bssopenapi20171214.QueryInstanceBillResponseBodyDataItemsItem, 0)
	resp = append(resp, result.Body.Data.Items.Item...)

	if *result.Body.Data.TotalCount > (ac.pageSize * ac.pageNum) {
		ac.pageNum++
		r, _ := ac.GetResourceMonthCost(billingCycle)
		resp = append(resp, r...)
	}

	return resp, err
}

func (ac *aliClint) GetResourceDayCost(billingCycle string) ([]ResourceCost, error) {
	return nil, nil
}

func (ac *aliClint) DownloadMonthBill(billingCycle string) (map[string]ResourceCost, []string, error) {
	billData, err := ac.GetResourceMonthCost(billingCycle)
	if err != nil {
		return nil, nil, err
	}

	insertDate := make([]models.BillData, 0, len(billData))
	resp := make(map[string]ResourceCost)
	resourceIds := make([]string, 0, len(billData))
	for _, v := range billData {
		resourceIds = append(resourceIds, *v.InstanceID)

		m := structs.Map(&v)
		insertDate = append(insertDate, models.BillData{
			Provider:   ac.provider,
			InstanceId: *v.InstanceID,
			Attrs:      models.ResAttrs(m),
		})

		resp[*v.InstanceID] = ResourceCost{
			ProductCode:    *v.ProductCode,
			InstanceId:     *v.InstanceID,
			InstanceConfig: *v.InstanceConfig,
			PretaxAmount:   *v.PretaxAmount,
			Region:         *v.Region,
			Currency:       *v.Currency,
			Cycle:          billingCycle,
			Provider:       ac.provider,
		}
	}

	if err := db.Get().Insert(&insertDate); err != nil {
		return nil, nil, err
	}

	return resp, resourceIds, nil
}

func (ac *aliClint) ParseBill(data []ResourceCost, resourceIds []string) ([]BillData, error) {
	return nil, nil
}
