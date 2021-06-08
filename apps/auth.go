package apps

import (
	"cloudiac/consts/e"
	"cloudiac/libs/ctx"
	"cloudiac/models/forms"
	"cloudiac/services"
	"cloudiac/utils"
	"fmt"
	"net/http"
	"time"
)

type Data struct {
	//UserInfo *models.User
	Token    string  `json:"token"`
}

func Login(c *ctx.ServiceCtx, form *forms.LoginForm) (resp interface{}, er e.Error) {
	c.AddLogField("action", fmt.Sprintf("user login: %s", form.Email))

	user, err := services.GetUserByEmail(c.DB(), form.Email)
	if err != nil {
		if e.IsRecordNotFound(err) {
			// 找不到账号时也返回 InvalidPassword 错误，避免暴露系统中己有用户账号
			return nil, e.New(e.InvalidPassword, http.StatusBadRequest)
		}
		return nil, e.New(e.DBError, err)
	}

	valid, err := utils.CheckPassword(form.Password, user.Password)
	if err != nil {
		return nil, e.New(e.ValidateError, http.StatusInternalServerError, err)
	}
	if !valid {
		return nil, e.New(e.InvalidPassword, http.StatusBadRequest)
	}

	token, err := services.GenerateToken(user.Id, user.Name, user.IsAdmin, 1*24*time.Hour)
	if err != nil {
		c.Logger().Errorf("name [%s] generateToken error: %v", user.Email, err)
		return nil, e.New(e.InvalidPassword, http.StatusBadRequest)
	}

	data := Data{
		//UserInfo: user,
		Token: token,
	}

	return data, nil
}