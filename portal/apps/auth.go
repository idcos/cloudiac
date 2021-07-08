package apps

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
	"cloudiac/utils"
	"cloudiac/utils/mail"
	"fmt"
	"net/http"
	"time"
)

// Login 用户登陆
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

	data := models.LoginResp{
		//UserInfo: user,
		Token: token,
	}

	return data, nil
}

// UserPassReset 用户重置密码
func UserPassReset(c *ctx.ServiceCtx, form *forms.DetailUserForm) (*models.User, e.Error) {
	initPass := utils.GenPasswd(6, "mix")
	hashedPassword, err := services.HashPassword(initPass)
	if err != nil {
		c.Logger().Errorf("error hash password %s", err)
		return nil, err
	}

	attrs := models.Attrs{}
	attrs["password"] = hashedPassword

	user, err := services.UpdateUser(c.DB(), form.Id, attrs)

	resp := struct {
		*models.User
		InitPass string
	}{
		User:     user,
		InitPass: initPass,
	}

	// TODO: 需确定邮件内容
	go func() {
		err := mail.SendMail([]string{user.Email}, emailSubjectResetPassword, utils.SprintTemplate(emailBodyResetPassword, resp))
		if err != nil {
			c.Logger().Errorf("error send mail to %s, err %s", user.Email, err)
		}
	}()

	return user, err
}
