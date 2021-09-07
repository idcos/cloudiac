// Copyright 2021 CloudJ Company Limited. All rights reserved.

package main

import (
	"cloudiac/configs"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/portal/services"
	"cloudiac/utils"
	"fmt"
)

// ./iac-tool password --email admin@example.com new_password

type ChangePassword struct {
	Email string `long:"email" description:"user email" required:"true"`
}

func (*ChangePassword) Usage() string {
	return `<new password>`
}

func (p *ChangePassword) Execute(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("new password is required")
	}

	configs.Init(opt.Config)
	db.Init(configs.Get().Mysql)

	password := args[0]
	user, err := services.GetUserByEmail(db.Get(), p.Email)
	if err != nil {
		if err.Code() == e.UserNotExists {
			return fmt.Errorf("user not exists")
		}
		return err
	}

	logger.Infof("update user password, email=%s, id=%s", user.Email, user.Id)
	hashedPass, er := utils.HashPassword(password)
	if er != nil {
		return er
	}

	_, err = services.UpdateUser(db.Get(), user.Id, models.Attrs{"password": hashedPass})
	return err
}
