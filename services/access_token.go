package services

import (
	"cloudiac/libs/db"
	"cloudiac/models"
)

func AccessTokenHandlerQuery(query *db.Session, accessToken string) *db.Session {
	return query.Table(models.TemplateAccessToken{}.TableName()).
		Where("access_token = ?", accessToken).
		Joins("left join iac_template as tpl on tpl.guid = iac_template_access_token.tpl_guid").
		LazySelectAppend("tpl.*").
		LazySelectAppend("iac_template_access_token.action,iac_template_access_token.access_token,iac_template_access_token.tpl_guid")
}
