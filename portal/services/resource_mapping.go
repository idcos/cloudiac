package services

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"fmt"
)

// SearchResourceMappingExpress 查询映射表达式
func SearchResourceMappingExpress(tx *db.Session, rmc []*models.ResourceMappingCondition) (map[string]string, e.Error) {
	if len(rmc) == 0 {
		return map[string]string{}, nil
	}
	// 条件去重
	rmcMap := map[string]bool{}
	newRmc := make([]*models.ResourceMappingCondition, 0)
	for _, item := range rmc {
		key := buildResourceMappingMapKey(item.Provider, item.Type, item.Code)
		if _, ok := rmcMap[key]; ok {
			continue
		} else {
			rmcMap[key] = true
			newRmc = append(newRmc, item)
		}
	}
	query := tx.Model(&models.ResourceMapping{})
	for _, item := range newRmc {
		query = query.Or("provider = ? and type = ? and code = ?", item.Provider, item.Type, item.Code)
	}
	rms := make([]*models.ResourceMapping, 0)
	if err := query.Scan(&rms); err != nil {
		return nil, e.New(e.DBError, err)
	}
	res := map[string]string{}
	for _, item := range rms {
		key := buildResourceMappingMapKey(item.Provider, item.Type, item.Code)
		res[key] = item.Express
	}
	return res, nil
}

func buildResourceMappingMapKey(provider, rmType, code string) string {
	return fmt.Sprintf("%s_%s_%s", provider, rmType, code)
}
