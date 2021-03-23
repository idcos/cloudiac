package apps

import (
	"cloudiac/consts/e"
	"cloudiac/libs/ctx"
	"cloudiac/libs/page"
	"cloudiac/models"
	"cloudiac/models/forms"
	"cloudiac/services"
	"cloudiac/utils"
	"fmt"
	"net/http"
)

func CreateUser(c *ctx.ServiceCtx, form *forms.CreateUserForm) (*models.User, e.Error) {
	c.AddLogField("action", fmt.Sprintf("create user %s", form.Name))

	initPass := utils.RandomStr(6)
	hashedPassword, er := services.HashPassword(initPass)
	if er != nil {
		return nil, er
	}

	tx := c.Tx().Debug()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	user, err := func() (*models.User, e.Error) {
		var (
			user *models.User
			err    e.Error
			er     e.Error
		)

		user, err = services.CreateUser(tx, models.User{
			Name:     form.Name,
			Password: hashedPassword,
			Phone:    form.Phone,
			Email:    form.Email,
			InitPass: initPass,
		})
		if err != nil {
			return nil, err
		}

		// 建立用户与组织间关联
		_, er = services.CreateUserOrgMap(tx, models.UserOrgMap{
			OrgId: c.OrgId,
			UserId: user.Id,
		})
		if er != nil {
			return nil, er
		}

		return user, nil
	}()
	if err != nil {
		_ = tx.Rollback()
		return nil, er
	}

	if err := tx.Commit(); err != nil {
		return nil, e.New(e.DBError, err)
	}

	return user, nil
}

type searchUserResp struct {
	models.User
	Password    string `json:"-"`
	InitPass    string `json:"-"`
}

func SearchUser(c *ctx.ServiceCtx, form *forms.SearchUserForm) (interface{}, e.Error) {
	query := services.QueryUser(c.DB())
	if form.Status != "" {
		query = query.Where("status = ?", form.Status)
	}
	if form.Q != "" {
		qs := "%" + form.Q + "%"
		query = query.Where("name LIKE ? OR phone LIKE ? OR email LIKE ? ", qs, qs, qs)
	}

	query = query.Order("created_at DESC")
	p := page.New(form.CurrentPage(), form.PageSize(), query)
	users := make([]*searchUserResp, 0)
	if err := p.Scan(&users); err != nil {
		return nil, e.New(e.DBError, err)
	}

	return page.PageResp{
		Total:    p.MustTotal(),
		PageSize: p.Size,
		List:     users,
	}, nil
}

func UpdateUser(c *ctx.ServiceCtx, form *forms.UpdateUserForm) (user *models.User, err e.Error) {
	c.AddLogField("action", fmt.Sprintf("update user %d", form.Id))
	if form.Id == 0 {
		return nil, e.New(e.BadRequest, fmt.Errorf("missing 'id'"))
	}

	attrs := models.Attrs{}
	if form.HasKey("name") {
		attrs["name"] = form.Name
	}

	if form.HasKey("phone") {
		attrs["phone"] = form.Phone
	}

	if form.HasKey("oldPassword") {
		if !form.HasKey("newPassword") {
			return nil, e.New(e.BadParam, http.StatusBadRequest)
		}
		user, er := services.GetUserById(c.DB(), form.Id)
		if er != nil {
			return nil, er
		}

		valid, err := utils.CheckPassword(form.OldPassword, user.Password)
		if err != nil {
			return nil, e.New(e.DBError, http.StatusInternalServerError, err)
		}
		if !valid {
			return nil, e.New(e.InvalidPassword, http.StatusBadRequest)
		}

		newPassword, er := services.HashPassword(form.NewPassword)
		if er != nil {
			return nil, er
		}
		attrs["password"] = newPassword
	}

	user, err = services.UpdateUser(c.DB(), form.Id, attrs)
	return
}

func DeleteUserOrgMap(c *ctx.ServiceCtx, form *forms.DeleteUserForm) (result interface{}, re e.Error) {
	c.AddLogField("action", fmt.Sprintf("delete user %d for org %d", form.Id, c.OrgId))

	tx := c.Tx()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := services.DeleteUserOrgMap(tx, form.Id, c.OrgId); err != nil {
		tx.Rollback()
		return nil, err
	} else if err := tx.Commit(); err != nil {
		return nil, e.New(e.DBError, err)
	}
	c.Logger().Infof("delete user ", form.Id, " for org ", c.OrgId, " succeed")

	return
}

func UserPassReset(c *ctx.ServiceCtx, form *forms.DetailUserForm) (user *models.User, err e.Error) {
	initPass := utils.RandomStr(6)
	hashedPassword, _ := services.HashPassword(initPass)

	attrs := models.Attrs{}
	attrs["init_pass"] = initPass
	attrs["password"] = hashedPassword

	user, err = services.UpdateUser(c.DB(), form.Id, attrs)
	return
}

func UserDetail(c *ctx.ServiceCtx, form *forms.DetailUserForm) (resp interface{}, er e.Error) {
	user, err := services.GetUserById(c.DB(), form.Id)
	if err != nil {
		return nil, e.New(e.DBError, http.StatusInternalServerError, err)
	}
	return user, nil
}
