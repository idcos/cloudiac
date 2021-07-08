package services

import (
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
)

func GetEnv(sess *db.Session, id models.Id) (*models.Env, error) {
	env := models.Env{}
	err := sess.Where("id = ?", id).First(&env)
	return &env, err
}
