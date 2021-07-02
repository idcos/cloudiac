package apps

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/libs/page"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
	"cloudiac/utils"
	"encoding/json"
	"fmt"
	"net/http"
)

// CreateUser 创建用户
func CreateUser(c *ctx.ServiceCtx, form *forms.CreateUserForm) (*models.User, e.Error) {
	c.AddLogField("action", fmt.Sprintf("create user %s", form.Name))

	initPass := utils.GenPasswd(6, "mix")
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
			err  e.Error
			er   e.Error
		)

		user, err = services.CreateUser(tx, models.User{
			Name:     form.Name,
			Password: hashedPassword,
			Phone:    form.Phone,
			Email:    form.Email,
		})
		if err != nil {
			if e.IsDuplicate(err) {
				user, _ = services.GetUserByEmail(tx, form.Email)
			} else {
				return nil, err
			}
		}

		// 建立用户与组织间关联
		_, er = services.CreateUserOrgRel(tx, models.UserOrg{
			OrgId:  c.OrgId,
			UserId: user.Id,
		})
		if er != nil {
			return nil, er
		}

		return user, nil
	}()
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, e.New(e.DBError, err)
	}

	return user, nil
}

type searchUserResp struct {
	models.User
	Password string `json:"-"`
	InitPass string `json:"-"`
	Role     string `json:"role"`
}

// SearchUser 查询用户列表
func SearchUser(c *ctx.ServiceCtx, form *forms.SearchUserForm) (interface{}, e.Error) {
	userOrgRel, err := services.GetUserByOrg(c.DB(), c.OrgId)
	var userIds []models.Id
	for _, o := range userOrgRel {
		userIds = append(userIds, o.UserId)
	}
	if err != nil {
		return nil, e.New(e.DBError, err)
	}
	query := services.QueryUser(c.DB())
	query = query.Where("id in (?)", userIds)
	if form.Status != "" {
		query = query.Where("status = ?", form.Status)
	}
	if form.Q != "" {
		qs := "%" + form.Q + "%"
		query = query.Where("name LIKE ? OR phone LIKE ? OR email LIKE ? ", qs, qs, qs)
	}

	if form.SortField() == "" {
		query = query.Order("created_at DESC")
	}
	p := page.New(form.CurrentPage(), form.PageSize(), query)
	users := make([]*searchUserResp, 0)
	if err := p.Scan(&users); err != nil {
		return nil, e.New(e.DBError, err)
	}
	for _, user := range users {
		for _, org := range userOrgRel {
			if user.Id == org.UserId {
				user.Role = org.Role
				break
			}
		}
	}

	return page.PageResp{
		Total:    p.MustTotal(),
		PageSize: p.Size,
		List:     users,
	}, nil
}

// UpdateUser 用户信息编辑
func UpdateUser(c *ctx.ServiceCtx, form *forms.UpdateUserForm) (user *models.User, err e.Error) {
	c.AddLogField("action", fmt.Sprintf("update user %d", form.Id))
	if form.Id == "" {
		return nil, e.New(e.BadRequest, fmt.Errorf("missing 'id'"))
	}
	if c.IsSuperAdmin == false && c.Role != "owner" && c.UserId != form.Id {
		return nil, e.New(e.PermissionDeny, http.StatusForbidden)
	}
	if c.UserId != form.Id {
		userOrgRel, _ := services.FindUsersOrgRel(c.DB(), form.Id, c.OrgId)
		if len(userOrgRel) == 0 {
			return nil, e.New(e.PermissionDeny, http.StatusForbidden)
		}
	}

	attrs := models.Attrs{}
	if form.HasKey("name") {
		attrs["name"] = form.Name
	}

	if form.HasKey("phone") {
		attrs["phone"] = form.Phone
	}

	if form.HasKey("newbieGuide") {
		b, _ := json.Marshal(form.NewbieGuide)
		attrs["newbie_guide"] = b
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

	return services.UpdateUser(c.DB().Debug(), form.Id, attrs)
}

// UserDetail 获取单个用户详情
func UserDetail(c *ctx.ServiceCtx, id models.Id) (*models.User, e.Error) {
	user, err := services.GetUserById(c.DB(), id)
	if err != nil && err.Code() == e.UserNotExists {
		// 通过 /auth/me 或者 /users/:userId 访问
		return nil, e.New(err.Code(), err, http.StatusNotFound)
	} else if err != nil {
		c.Logger().Errorf("error get user by id, err %s", err)
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	}
	return user, nil
}

// DeleteUser 删除用户
func DeleteUser(c *ctx.ServiceCtx, form *forms.DeleteUserForm) (interface{}, e.Error) {
	c.AddLogField("action", fmt.Sprintf("delete user %s", form.Id))
	// TODO
	return nil, nil
}
