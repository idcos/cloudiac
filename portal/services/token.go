// Copyright 2021 CloudJ Company Limited. All rights reserved.

package services

import (
	"cloudiac/configs"
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"time"
)

type Claims struct {
	UserId   models.Id `json:"userId"`
	Username string    `json:"username"`
	IsAdmin  bool      `json:"isAdmin"`
	jwt.StandardClaims
}

func GenerateToken(uid models.Id, name string, isAdmin bool, expireDuration time.Duration) (string, error) {
	expire := time.Now().Add(expireDuration)

	// 将 userId，姓名, 过期时间写入 token 中
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserId:   uid,
		Username: name,
		IsAdmin:  isAdmin,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expire.Unix(),
		},
	})

	return token.SignedString([]byte(configs.Get().JwtSecretKey))
}

func CreateToken(tx *db.Session, token models.Token) (*models.Token, e.Error) {
	if token.Id == "" {
		token.Id = models.NewId("t")
	}
	if err := models.Create(tx, &token); err != nil {
		if e.IsDuplicate(err) {
			return nil, e.New(e.TokenAlreadyExists, err)
		}
		return nil, e.New(e.DBError, err)
	}

	return &token, nil
}

func UpdateToken(tx *db.Session, id models.Id, attrs models.Attrs) (token *models.Token, er e.Error) {
	token = &models.Token{}
	if _, err := models.UpdateAttr(tx.Where("id = ?", id), &models.Token{}, attrs); err != nil {
		if e.IsDuplicate(err) {
			return nil, e.New(e.TokenAliasDuplicate)
		}
		return nil, e.New(e.DBError, fmt.Errorf("update token error: %v", err))
	}
	if err := tx.Where("id = ?", id).First(token); err != nil {
		return nil, e.New(e.DBError, fmt.Errorf("query token error: %v", err))
	}
	return
}

func QueryToken(query *db.Session, tokenType string) *db.Session {
	query = query.Model(&models.Token{})
	if tokenType != "" {
		query = query.Where("type = ?", tokenType)
	}
	return query
}

func DeleteToken(tx *db.Session, id models.Id) e.Error {
	if _, err := tx.Where("id = ?", id).Delete(&models.Token{}); err != nil {
		return e.New(e.DBError, fmt.Errorf("delete token error: %v", err))
	}
	return nil
}

func TokenExists(query *db.Session, apiToken string) (bool, *models.Token) {
	token := &models.Token{}
	q := query.Model(&models.Token{}).
		Where("token = ?", apiToken).
		Where("status = 'enable'")
	exists, err := q.Exists()
	if err != nil {
		return exists, nil
	}
	if err := q.First(token); err != nil {
		return exists, nil
	}

	return exists, token
}

func DetailTriggerToken(dbSess *db.Session, orgId, envId models.Id, action string) (interface{}, e.Error) {
	token := models.Token{}
	query := QueryToken(dbSess.Where("org_id = ?", orgId).
		Where("env_id = ?", envId).
		Where("action = ?", action), consts.TokenTrigger)
	if err := query.First(&token); err != nil {
		if e.IsRecordNotFound(err) {
			return nil, e.New(e.TokenNotExists)
		}
		return nil, e.New(e.DBError, err)
	}
	return token, nil
}

func IsExistsTriggerToken(dbSess *db.Session, tokenTrigger string) (*models.Token, e.Error) {
	token := models.Token{}
	if err := dbSess.
		Table(models.Token{}.TableName()).
		Where("`key` = ?", tokenTrigger).
		Where("`type` = ?", consts.TokenTrigger).
		First(&token); err != nil {
		if e.IsRecordNotFound(err) {
			return nil, e.New(e.TokenNotExists)
		}
		return nil, e.New(e.DBError, err)
	}
	return &token, nil
}
