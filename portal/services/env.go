package services

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
)

func GetEnv(sess *db.Session, id models.Id) (*models.Env, error) {
	env := models.Env{}
	err := sess.Where("id = ?", id).First(&env)
	return &env, err
}

func GetEnvByTplId(tx *db.Session, id models.Id) ([]models.Env, error){
	env := []models.Env{}
	if err := tx.Where("tplId = ?", id).Find(&env); err != nil {
		return nil, e.New(e.DBError, err)
	}
	return env, nil
}
