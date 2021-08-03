// Copyright 2021 CloudJ Company Limited. All rights reserved.

package services

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/utils"
	"fmt"
	"github.com/sirupsen/logrus"
)

func CreateKey(tx *db.Session, key models.Key) (*models.Key, e.Error) {
	if key.Id == "" {
		key.Id = models.NewId("k")
	}
	if err := models.Create(tx, &key); err != nil {
		if e.IsDuplicate(err) {
			return nil, e.New(e.KeyAlreadyExists, err)
		}
		return nil, e.New(e.DBError, err)
	}

	return &key, nil
}

func UpdateKey(tx *db.Session, id models.Id, attrs models.Attrs) (key *models.Key, er e.Error) {
	key = &models.Key{}
	if aff, err := models.UpdateAttr(tx.Where("id = ?", id), &models.Key{}, attrs); err != nil {
		if e.IsDuplicate(err) {
			return nil, e.New(e.KeyAliasDuplicate)
		} else if e.IsRecordNotFound(err) {
			return nil, e.New(e.KeyNotExist)
		}
		return nil, e.New(e.DBError, fmt.Errorf("update key error: %v", err))
	} else {
		if aff == 0 {
			return nil, e.New(e.KeyNotExist)
		}
	}
	if err := tx.Where("id = ?", id).First(key); err != nil {
		logrus.Errorf("query %s", tx.Expr())
		return nil, e.New(e.DBError, fmt.Errorf("query key error: %v", err))
	}

	return
}

func QueryKey(query *db.Session) *db.Session {
	query = query.Model(&models.Key{})
	return query
}

func DeleteKey(tx *db.Session, id models.Id) e.Error {
	if _, err := tx.Where("id = ?", id).Delete(&models.Key{}); err != nil {
		if e.IsRecordNotFound(err) {
			return e.New(e.KeyNotExist)
		}
		return e.New(e.DBError, fmt.Errorf("delete key error: %v", err))
	}
	return nil
}

// GetKeyById 查询密钥详情
// decrypt bool 是否自动解密为原始密钥
func GetKeyById(query *db.Session, id models.Id, decrypt bool) (*models.Key, e.Error) {
	key := models.Key{}
	if err := query.Model(models.Key{}).Where("id = ?", id).First(&key); err != nil {
		if e.IsRecordNotFound(err) {
			return nil, e.New(e.KeyNotExist)
		}
		return nil, e.New(e.DBError, err)
	}
	if decrypt {
		var err error
		key.Content, err = utils.AesDecrypt(key.Content)
		if err != nil {
			return nil, e.New(e.KeyDecryptFail, err)
		}
	}
	return &key, nil
}
