package apps

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
	"cloudiac/utils"
	"fmt"
	"net/http"
	"time"
)

type LoginResp struct {
	//UserInfo *models.User
	Token string `json:"token"`
}

// Login 用户登陆
// @Tags 鉴权
// @Description 用户登陆接口
// @Accept multipart/form-data
// @Accept json
// @Param body formData forms.LoginForm true "parameter"
// @router /api/v1/auth/login [post]
// @Success 200 {object} LoginResp
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

	data := LoginResp{
		//UserInfo: user,
		Token: token,
	}

	return data, nil
}
