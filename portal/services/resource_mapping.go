package services

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"fmt"
)

// SearchResourceMappingExpress 查询映射表达式
func SearchResourceMappingExpress(tx *db.Session, rmc []*models.ResourceMappingCondition) (map[string]string, e.Error) {
	query := tx
	for _, item := range rmc {
		query.Or("provider = ? and type = ? and code = ?", item.Provider, item.Type, item.Code)
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
