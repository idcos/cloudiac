// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package services

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/portal/services/billcollect"
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

func ParseBill(resCost map[string]billcollect.ResourceCost, res []models.Resource, vgId models.Id) ([]models.Bill, []string) {
	resIds := make([]string, 0)
	resp := make([]models.Bill, 0)
	for _, v := range res {
		if _, ok := resCost[v.ResId.String()]; ok {
			resIds = append(resIds, v.Id.String())
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

// ProjectEnabledBill 检查项目是否启用了计费
func ProjectEnabledBill(sess *db.Session, pid models.Id) (bool, e.Error) {
	//select var_group_id from iac_variable_group_project_rel where project_id in ('', 'p-c8pkc3bn6m84mseaclrg')
	varGroupIdQuery := sess.Model(&models.VariableGroupProjectRel{}).
		Where("project_id IN ('', ?)", pid).
		Select("var_group_id")

	ok, err := sess.Model(&models.VariableGroup{}).
		Where("id in (?) and cost_counted", varGroupIdQuery.Expr()).
		Exists()
	if err != nil {
		return false, e.New(e.DBError, err)
	}
	return ok, nil
}
