// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package services

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/portal/services/billcollect"
	"cloudiac/utils/logs"
)

func GetVgByBillConf(dbSess *db.Session) ([]models.VariableGroup, e.Error) {
	resp := make([]models.VariableGroup, 0)
	if err := dbSess.Model(models.VariableGroup{}).
		Where("cost_counted = ?", true).
		Find(&resp); err != nil {
		return nil, e.New(e.DBError, err)
	}

	return resp, nil
}

func GetVgByOrgId(dbSess *db.Session, orgId string) ([]models.VariableGroup, e.Error) {
	resp := make([]models.VariableGroup, 0)
	if err := dbSess.Model(models.VariableGroup{}).
		Where("cost_counted = ?", true).
		Where("org_id = ?", orgId).
		Find(&resp); err != nil {
		return nil, e.New(e.DBError, err)
	}

	return resp, nil
}

func GetVgByProjectId(dbSess *db.Session, projectId string) ([]models.VariableGroup, e.Error) {
	resp := make([]models.VariableGroup, 0)
	varGroupIdQuery := dbSess.Model(&models.VariableGroupProjectRel{}).
		Where("project_id IN (?)", projectId).
		Select("var_group_id")

	if err := dbSess.Model(&models.VariableGroup{}).
		Where("id in (?) and cost_counted", varGroupIdQuery.Expr()).Find(&resp); err != nil {
		return nil, e.New(e.DBError, err)
	}

	return resp, nil
}

func GetVgById(dbSess *db.Session, id string) ([]models.VariableGroup, e.Error) {
	resp := make([]models.VariableGroup, 0)
	if err := dbSess.Model(models.VariableGroup{}).
		Where("id = ?", id).
		Find(&resp); err != nil {
		return nil, e.New(e.DBError, err)
	}

	return resp, nil
}

func BuildBillData(resCost map[string]billcollect.ResourceCost, res []models.Resource, vgId models.Id) ([]models.Bill, []string) {
	resIds := make([]string, 0)
	resp := make([]models.Bill, 0)
	for _, v := range res {
		if _, ok := resCost[v.ResId.String()]; ok {
			resIds = append(resIds, v.ResId.String())
			resp = append(resp, models.Bill{
				OrgId:          v.OrgId,
				ProjectId:      v.ProjectId,
				EnvId:          v.EnvId,
				VgId:           vgId,
				ProductCode:    resCost[v.ResId.String()].ProductCode,
				InstanceId:     resCost[v.ResId.String()].InstanceId,
				InstanceConfig: resCost[v.ResId.String()].InstanceConfig,
				PretaxAmount:   resCost[v.ResId.String()].PretaxAmount,
				Region:         resCost[v.ResId.String()].Region,
				Currency:       resCost[v.ResId.String()].Currency,
				Cycle:          resCost[v.ResId.String()].Cycle,
				Provider:       resCost[v.ResId.String()].Provider,
			})
		}
	}

	return resp, resIds
}

func DeleteResourceBill(dbSess *db.Session, resIdS []string, cycle string) error {
	if _, err := dbSess.Where("instance_id in (?)", resIdS).
		Where("cycle = ?", cycle).
		Delete(models.Bill{}); err != nil {
		return err
	}
	return nil
}

func BuildVgBilling(tx *db.Session, vg models.VariableGroup, lg logs.Logger, billingCycle string) {
	// 获取账单provider
	bp, err := billcollect.GetBillProvider(&vg)
	if err != nil {
		lg.Errorf("get bill provider failed vgId: %s, vgName: %s, provider: %s, err: %s", vg.Id, vg.Name, vg.Provider, err)
		return
	}

	lg.Infof("vgId: %s, vgName: %s, provider: %s -> GetBillProvider finish", vg.Id, vg.Name, vg.Provider)

	// 解析原始账单数据
	resCostAttr, resourceIds, insertDate, err := bp.ParseMonthBill(billingCycle)
	if err != nil {
		lg.Errorf("parse bill failed vgId: %s, vgName: %s, provider: %s, err: %s", vg.Id, vg.Name, vg.Provider, err)
		return
	}

	// 写入原始账单数据
	if err := tx.Insert(&insertDate); err != nil {
		return
	}

	lg.Infof("vgId: %s, vgName: %s, provider: %s -> ParseMonthBill finish", vg.Id, vg.Name, vg.Provider)

	// 查询资源账号关联的项目
	projectIds, err := GetProjectIdsByVgId(tx, vg.Id)
	if err != nil {
		lg.Errorf("query project ids failed vgId: %s, vgName: %s, provider: %s, err: %s", vg.Id, vg.Name, vg.Provider, err)
		return
	}

	// 查询iac resource数据
	res, err := GetResourceByIdsInProvider(tx, resourceIds, projectIds, vg)
	if err != nil {
		lg.Errorf("query iac resource failed vgId: %s, vgName: %s, provider: %s, err: %s", vg.Id, vg.Name, vg.Provider, err)
		return
	}

	// 解析账单数据，构建入库数据
	bills, resIds := BuildBillData(resCostAttr, res, vg.Id)
	if len(bills) == 0 {
		lg.Infof("resource not matched collect billing vgId: %s, vgName: %s, provider: %s", vg.Id, vg.Name, vg.Provider)
		return
	}

	lg.Infof("vgId: %s, vgName: %s, provider: %s -> ParseBill finish", vg.Id, vg.Name, vg.Provider)

	// 删除上次采集的数据
	if err := DeleteResourceBill(tx, resIds, billingCycle); err != nil {
		lg.Errorf("del last bill data failed vgId: %s, vgName: %s, provider: %s, err: %s", vg.Id, vg.Name, vg.Provider, err)
		return
	}

	lg.Infof("vgId: %s, vgName: %s, provider: %s -> DeleteResourceBill finish", vg.Id, vg.Name, vg.Provider)

	if err := tx.Insert(bills); err != nil {
		lg.Errorf("bill insert failed vgId: %s, vgName: %s, provider: %s, err: %s", vg.Id, vg.Name, vg.Provider, err)
		return
	}

	lg.Infof("vgId: %s, vgName: %s, provider: %s -> InsertResourceBill finish", vg.Id, vg.Name, vg.Provider)
}
