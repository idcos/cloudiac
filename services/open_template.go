package services

import (
	"cloudiac/libs/db"
	"cloudiac/models"
)

func OpenSearchTemplate(tx *db.Session) *db.Session {
	return tx.Table(models.Template{}.TableName()).
		Where("status = 'enable'").
		LazySelectAppend("guid", "name")
}
