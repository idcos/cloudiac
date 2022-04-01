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

func ParseBill(resCost map[string]billcollect.ResourceCost, res []models.Resource, vgId string) []models.Bill {
	resp := make([]models.Bill, 0)
	for _, v := range res {
		if _, ok := resCost[v.ResId.String()]; ok {
			resp = append(resp, models.Bill{
				OrgId:          v.OrgId,
				ProjectId:      v.ProjectId,
				VgId:           models.Id(vgId),
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

	return resp
}

func DeleteResourceBill(dbSess *db.Session, resIdS []string, cycle string) error {
	if _, err := dbSess.Where("instance_id in (?)", resIdS).
		Where("cycle = ?", cycle).
		Delete(models.Bill{}); err != nil {
		return err
	}
	return nil
}
