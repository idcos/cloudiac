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

func GetResourceByVg(dbSess *db.Session, vg models.VariableGroup) ([]models.Resource, e.Error) {
	resp := make([]models.Resource, 0)
	if err := dbSess.Model(models.VariableGroup{}).
		Where("cost_counted = ?", true).
		Find(&resp); err != nil {
		return nil, e.New(e.DBError, err)
	}
	return resp, nil
}

func BillData(vg models.VariableGroup,billingCycle string) ([]billcollect.ResourceCost, e.Error) {
	billInstance, err := billcollect.GetBillInstance(&vg)
	if err != nil {
		return nil, err
	}
	cline, err := billInstance.Clint()
	if err != nil {
		return nil, err
	}

	resp, err := cline.GetResourceMonthCost(billingCycle)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
