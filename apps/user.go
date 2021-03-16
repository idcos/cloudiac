package apps

// 用户管理示例

import (
	"fmt"
	"net/http"
	"cloudiac/libs/db"
	"strings"

	"cloudiac/consts/e"
	"cloudiac/libs/ctx"
	"cloudiac/libs/page"
	"cloudiac/models"
	"cloudiac/models/forms"
	"cloudiac/services"
	"cloudiac/utils"
)

// 创建用户功能未开放，所以该函数还未做鉴权
func CreateUser(c *ctx.ServiceCtx, form *forms.CreateUserForm) (*models.User, e.Error) {
	c.AddLogField("action", fmt.Sprintf("create user %s", form.UserName))

	hashedPassword, er := services.HashPassword(form.Password)
	if er != nil {
		return nil, er
	}
	user, err := services.CreateUser(c.DB(), form.TenantId, models.User{
		Username: form.UserName,
		Password: hashedPassword,
		Phone:    form.Phone,
		Email:    form.Email,
	}, false)
	if err != nil {
		return nil, err
	}
	return user, nil
}

type searchUserResp struct {
	models.User
	CustomerIds   string `json:"-"`
	CustomerNames string `json:"-"`

	Customers []string `json:"customers" gorm:"-"`
}

func SearchUser(c *ctx.ServiceCtx, form *forms.SearchUserForm) (interface{}, e.Error) {
	query := services.QueryUser(c.DB())
	if form.Status != 0 {
		query = query.Where("status = ?", form.Status)
	}
	if form.Q != "" {
		qs := "%" + form.Q + "%"
		query = query.Where("username LIKE ? OR phone LIKE ? OR email LIKE ? "+
			"OR customer_names LIKE ?", qs, qs, qs, qs)
	}

	query = query.Order("created_at DESC")
	p := page.New(form.CurrentPage(), form.PageSize(), query)
	users := make([]*searchUserResp, 0)
	if err := p.Scan(&users); err != nil {
		return nil, e.New(e.DBError, err)
	}

	for _, u := range users {
		if u.CustomerNames != "" {
			u.Customers = strings.Split(u.CustomerNames, ",")
		} else {
			u.Customers = make([]string, 0)
		}
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
	if form.HasKey("username") {
		attrs["username"] = form.UserName
	}

	if form.HasKey("phone") {
		attrs["phone"] = form.Phone
	}

	// 邮箱不可编辑
	//if form.HasKey("email") {
	//	attrs["email"] = form.Email
	//}

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

// 用户删除功能暂未开放，所以还未做鉴权
func DeleteUser(c *ctx.ServiceCtx, form *forms.DeleteUserForm) (result interface{}, re e.Error) {
	c.AddLogField("action", fmt.Sprintf("delete user %d", form.Id))

	tx := c.Tx()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := services.DeleteUser(tx, form.Id); err != nil {
		tx.Rollback()
		return nil, err
	} else if err := tx.Commit(); err != nil {
		return nil, e.New(e.DBError, err)
	}
	c.Logger().Infof("delete user succeed")

	return
}

func DisableUser(c *ctx.ServiceCtx, form *forms.DisableUserForm) (interface{}, e.Error) {
	user, er := services.GetUserById(c.DB(), form.Id)
	if er != nil {
		return nil, er
	}

	if user.Status == form.Status {
		return user, nil
	} else if user.Status != models.UserStatusNormal && user.Status != models.UserStatusDisabled {
		return nil, services.CheckUserStatus(user.Status)
	}

	user, err := services.UpdateUser(c.DB(), form.Id, models.Attrs{"status": form.Status})
	if err != nil {
		return nil, err
	}
	// 在 Auth 中间件和登录认证时会检查用户的 status
	return user, nil
}

func UserDetail(c *ctx.ServiceCtx, form *forms.DetailUserForm) (resp interface{}, er e.Error) {
	user, err := services.GetUserById(c.DB(), form.Id)
	if err != nil {
		return nil, e.New(e.DBError, http.StatusInternalServerError, err)
	}
	return user, nil
}

// 内部调用用于添加用户的函数
func AddUser(sess *db.Session, tid uint, email string, password string,
	isAdmin bool, username string) (*models.User, error) {

	if tid == 0 {
		return nil, fmt.Errorf("tenant id must be specified")
	}

	hashedPass, err := services.HashPassword(password)
	if err != nil {
		return nil, err
	}

	if username == "" {
		idx := strings.Index(email, "@")
		if idx > 0 {
			username = email[0:idx]
		}
	}

	user, err := services.CreateUser(sess, tid, models.User{
		Username: username,
		Email:    email,
		Password: hashedPass,
		Phone:    "",
		Status:   models.UserStatusNormal,
	}, isAdmin)
	if err != nil {
		return nil, err
	}

	return user, nil
}
