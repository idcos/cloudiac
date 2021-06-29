package services

import (
	"cloudiac/libs/db"
	"cloudiac/models"
	"fmt"
)

func OpenSearchTemplate(tx *db.Session, q string) *db.Session {
	query := tx.Table(models.Template{}.TableName()).
		Where("status = 'enable'").
		Where("save_state = 0").
		LazySelectAppend("guid", "name", "tpl_type", "description")
	if q != "" {
		query = query.Where("name like ?", fmt.Sprintf("%%%s%%", q))
	}
	return query
}
