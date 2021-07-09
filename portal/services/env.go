package services

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"fmt"
)

func GetEnv(sess *db.Session, id models.Id) (*models.Env, error) {
	env := models.Env{}
	err := sess.Where("id = ?", id).First(&env)
	return &env, err
}

func CreateEnv(tx *db.Session, env models.Env) (*models.Env, e.Error) {
	if env.Id == "" {
		env.Id = models.NewId("env")
	}
	if env.StatePath == "" {
		env.StatePath = env.DefaultStatPath()
	}
	if err := models.Create(tx, &env); err != nil {
		if e.IsDuplicate(err) {
			return nil, e.New(e.EnvAlreadyExists, err)
		}
		return nil, e.New(e.DBError, err)
	}

	return &env, nil
}

func UpdateEnv(tx *db.Session, id models.Id, attrs models.Attrs) (env *models.Env, re e.Error) {
	env = &models.Env{}
	if _, err := models.UpdateAttr(tx.Where("id = ?", id), &models.Env{}, attrs); err != nil {
		if e.IsDuplicate(err) {
			return nil, e.New(e.EnvAliasDuplicate)
		}
		return nil, e.New(e.DBError, fmt.Errorf("update env error: %v", err))
	}
	if err := tx.Where("id = ?", id).First(env); err != nil {
		return nil, e.New(e.DBError, fmt.Errorf("query env error: %v", err))
	}
	return
}

func DeleteEnv(tx *db.Session, id models.Id) e.Error {
	if _, err := tx.Where("id = ?", id).Delete(&models.Env{}); err != nil {
		return e.New(e.DBError, fmt.Errorf("delete env error: %v", err))
	}
	return nil
}

func GetEnvById(tx *db.Session, id models.Id) (*models.Env, e.Error) {
	o := models.Env{}
	if err := tx.Where("id = ?", id).First(&o); err != nil {
		if e.IsRecordNotFound(err) {
			return nil, e.New(e.EnvNotExists, err)
		}
		return nil, e.New(e.DBError, err)
	}
	return &o, nil
}

func QueryEnvDetail(query *db.Session) *db.Session {
	query = query.Model(&models.Env{})

	// 模板名称
	query = query.Joins("left join iac_template as t on t.id = iac_env.tpl_id").
		LazySelectAppend("t.name as template_name,iac_env.*")
	// 创建人姓名
	query = query.Joins("left join iac_user as u on u.id = iac_env.creator_id").
		LazySelectAppend("u.name as creator,iac_env.*")
	// 资源数量统计
	query = query.Joins("left join (select count(*) as resource_count, env_id from iac_env_res group by env_id) as r on r.env_id = iac_env.id").
		LazySelectAppend("r.resource_count, iac_env.*")

	return query
}

func GetEnvDetailById(query *db.Session, id models.Id) (*models.EnvDetail, e.Error) {
	d := models.EnvDetail{}
	if err := query.Where("iac_env.id = ?", id).First(&d); err != nil {
		if e.IsRecordNotFound(err) {
			return nil, e.New(e.EnvNotExists, err)
		}
		return nil, e.New(e.DBError, err)
	}
	return &d, nil
}

//
//func GetEnvIdsByUser(query *db.Session, userId models.Id) (envIds []models.Id, err error) {
//	var userEnvRel []*models.UserEnv
//	if err := query.Where("user_id = ?", userId).Find(&userEnvRel); err != nil {
//		return nil, e.AutoNew(err, e.DBError)
//	}
//	for _, o := range userEnvRel {
//		envIds = append(envIds, o.EnvId)
//	}
//	return
//}
