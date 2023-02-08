// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package billcollect

import (
	"cloudiac/portal/consts"
	"cloudiac/portal/models"
	"fmt"
	bssopenapi20171214 "github.com/alibabacloud-go/bssopenapi-20171214/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/fatih/structs"
	"github.com/pkg/errors"
)

func NewAlicloudBillProvider(vg *models.VariableGroup) (*aliProvider, error) {
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

	return &aliProvider{
		client:   result,
		pageNum:  1,
		pageSize: 300,
		provider: vg.Provider,
		vg:       vg,
	}, err
}

type aliProvider struct {
	client   *bssopenapi20171214.Client
	provider string
	pageNum  int32
	pageSize int32
	vg       *models.VariableGroup
}

func (ap *aliProvider) Provider() string {
	return ap.provider
}

func (ap *aliProvider) GetResourceMonthCost(billingCycle string) ([]*bssopenapi20171214.QueryInstanceBillResponseBodyDataItemsItem, error) {
	queryInstanceBillRequest := &bssopenapi20171214.QueryInstanceBillRequest{
		//BillingCycle: tea.String("2022-03"),
		BillingCycle: tea.String(billingCycle),
		PageNum:      tea.Int32(ap.pageNum),
		//单次请求最大值300
		PageSize: tea.Int32(ap.pageSize),
	}

	result, err := ap.client.QueryInstanceBill(queryInstanceBillRequest)
	if err != nil {
		return nil, err
	}
	if !*result.Body.Success {
		return nil, errors.New("get %s bill error")
	}
	resp := make([]*bssopenapi20171214.QueryInstanceBillResponseBodyDataItemsItem, 0)
	resp = append(resp, result.Body.Data.Items.Item...)

	if *result.Body.Data.TotalCount > (ap.pageSize * ap.pageNum) {
		ap.pageNum++
		r, _ := ap.GetResourceMonthCost(billingCycle)
		resp = append(resp, r...)
	}

	return resp, err
}

func (ap *aliProvider) GetResourceDayCost(billingCycle string) ([]ResourceCost, error) {
	return nil, nil
}

func (ap *aliProvider) ParseMonthBill(billingCycle string) (map[string]ResourceCost, []string, []models.BillData, error) {
	billData, err := ap.GetResourceMonthCost(billingCycle)
	if err != nil {
		return nil, nil, nil, err
	}

	insertDate := make([]models.BillData, 0, len(billData))
	resp := make(map[string]ResourceCost)
	resourceIds := make([]string, 0, len(billData))
	for index, v := range billData {
		resourceIds = append(resourceIds, *v.InstanceID)
		m := structs.Map(&billData[index])
		insertDate = append(insertDate, models.BillData{
			Provider:   ap.provider,
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
			Provider:       ap.provider,
		}
	}

	return resp, resourceIds, insertDate, nil
}
