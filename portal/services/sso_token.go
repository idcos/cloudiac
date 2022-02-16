// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package services

import (
	"cloudiac/configs"
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type SsoTokenClaims struct {
	jwt.StandardClaims

	UserId models.Id `json:"userId"`
}

func GenerateSsoToken(uid models.Id, expireDuration time.Duration) (string, error) {
	expire := time.Now().Add(expireDuration)

	// 将 userId，姓名, 过期时间写入 token 中
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, SsoTokenClaims{
		UserId: uid,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expire.Unix(),
			Subject:   consts.JwtSubjectSsoCode,
		},
	})

	return token.SignedString([]byte(configs.Get().JwtSecretKey))
}

func VerifySsoToken(tx *db.Session, tokenStr string) (*models.User, e.Error) {
	token, err := jwt.ParseWithClaims(tokenStr, &SsoTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(configs.Get().JwtSecretKey), nil
	})

	if err != nil {
		return nil, e.New(e.InvalidToken, err)
	}

	// get user info from db
	if claims, ok := token.Claims.(*SsoTokenClaims); ok && token.Valid {
		// 根据用户ID获取用户信息
		var user = models.User{}
		if err = tx.Where("id = ?", claims.UserId).First(&user); err != nil {
			return nil, e.New(e.UserNotExists, err)
		}

		return &user, nil
	}

	return nil, e.New(e.InvalidToken, fmt.Errorf("SSO token 不存在或者过期"))
}
