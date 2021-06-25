package services

import (
	"cloudiac/libs/db"
	"cloudiac/models"
)

func OpenSearchTemplate(tx *db.Session) *db.Session {
	return tx.Table(models.Template{}.TableName()).
		Where("status = 'enable'").
		Where("save_state = 0").
		LazySelectAppend("guid", "name", "tpl_type", "description")
}
