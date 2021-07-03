package apps

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/libs/page"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
	"cloudiac/utils"
	"cloudiac/utils/mail"
	"encoding/json"
	"fmt"
	"net/http"
)

// CreateUserResp 创建用户返回结果，带上初始化的随机密码
type CreateUserResp struct {
	*models.User
	InitPass string `json:"initPass" example:"rANd0m"` // 初始化密码
}

var (
	emailSubjectCreateUser = "注册用户成功通知"
	emailBodyCreateUser    = "尊敬的 {{.Name}}：\n\n恭喜您完成注册 CloudIaC 服务。\n\n这是您的登录详细信息：\n\n登录名：\t{{.Email}}\n密码：\t{{.InitPass}}\n\n为了保障您的安全，请立即登陆您的账号并修改初始密码。"
)

// CreateUser 创建用户
func CreateUser(c *ctx.ServiceCtx, form *forms.CreateUserForm) (*CreateUserResp, e.Error) {
	c.AddLogField("action", fmt.Sprintf("create user %s", form.Name))

	initPass := utils.GenPasswd(6, "mix")
	hashedPassword, err := services.HashPassword(initPass)
	if err != nil {
		c.Logger().Errorf("error hash password, err %s", err)
		return nil, err
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
		)

		user, err = services.CreateUser(tx, models.User{
			Name:     form.Name,
			Password: hashedPassword,
			Phone:    form.Phone,
			Email:    form.Email,
		})
		if err != nil && err.Code() == e.UserAlreadyExists {
			return nil, e.New(err.Code(), err, http.StatusBadRequest)
		} else if err != nil {
			c.Logger().Errorf("error create user, err %s", err)
			return nil, err
		}

		// 建立用户与组织间关联
		_, err = services.CreateUserOrgRel(tx, models.UserOrg{
			OrgId:  c.OrgId,
			UserId: user.Id,
		})
		if err != nil {
			c.Logger().Errorf("error create user , err %s", err)
			return nil, err
		}

		return user, nil
	}()
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("error commit user, err %s", err)
		return nil, e.New(e.DBError, err)
	}

	// 返回用户信息和初始化密码
	resp := CreateUserResp{
		User:     user,
		InitPass: initPass,
	}

	// TODO: 需确定邮件内容
	go func() {
		err := mail.SendMail([]string{user.Email}, emailSubjectCreateUser, utils.SprintTemplate(emailBodyCreateUser, resp))
		if err != nil {
			c.Logger().Errorf("error send mail to %s, err %s", user.Email, err)
		}
	}()

	return &resp, nil
}

type UserWithRoleResp struct {
	models.User
	Password string `json:"-"`
	Role     string `json:"role" example:"member"` // 角色
}

// SearchUser 查询用户列表
func SearchUser(c *ctx.ServiceCtx, form *forms.SearchUserForm) (interface{}, e.Error) {
	userOrgRel, err := services.GetUsersByOrg(c.DB(), c.OrgId)
	if err != nil {
		c.Logger().Errorf("error get users by org, err %s", err)
		return nil, e.New(e.DBError, err)
	}
	var userIds []models.Id
	for _, o := range userOrgRel {
		userIds = append(userIds, o.UserId)
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
	users := make([]*UserWithRoleResp, 0)
	if err := p.Scan(&users); err != nil {
		c.Logger().Errorf("error get users, err %s", err)
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

//ChangeUserStatus 修改用户启用/禁用状态
func ChangeUserStatus(c *ctx.ServiceCtx, form *forms.DisableUserForm) (*models.User, e.Error) {
	c.AddLogField("action", fmt.Sprintf("change user status %s", form.Id))

	if form.Status != models.Enable && form.Status != models.Disable {
		return nil, e.New(e.UserInvalidStatus, http.StatusBadRequest)
	}

	user, err := services.GetUserById(c.DB(), form.Id)
	if err != nil && err.Code() == e.UserNotExists {
		return nil, e.New(err.Code(), err, http.StatusBadRequest)
	} else if err != nil {
		c.Logger().Errorf("error get user by id, err %s", err)
		return nil, err
	}

	if user.Status == form.Status {
		return user, nil
	}

	user, err = services.UpdateUser(c.DB(), form.Id, models.Attrs{"status": form.Status})
	if err != nil {
		c.Logger().Errorf("error update user, err %s", err)
		return nil, e.New(e.DBError, err)
	}

	return user, nil
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
// 需要组织管理员权限，如果用户拥有多个组织权限，管理员需要拥有所有相关组织权限。
func DeleteUser(c *ctx.ServiceCtx, form *forms.DeleteUserForm) (interface{}, e.Error) {
	c.AddLogField("action", fmt.Sprintf("delete user %s", form.Id))
	user, err := services.GetUserById(c.DB(), form.Id)
	c.Logger().Errorf("del user %s", form.Id)
	if err != nil && err.Code() == e.UserNotExists {
		return nil, e.New(err.Code(), err, http.StatusNotFound)
	} else if err != nil {
		c.Logger().Errorf("error get user by id, err %s", err)
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	}

	tx := c.Tx().Debug()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	// TODO: 判断管理员是否拥有所有关联组织管理员权限

	// 解除组织关系
	orgIds, er := services.GetOrgIdsByUser(c.DB(), c.UserId)
	if er != nil {
		c.Logger().Errorf("error get org id by user, err %s", er)
		return nil, e.New(e.DBError, er)
	}
	for _, orgId := range orgIds {
		if err := services.DeleteUserOrgRel(tx, user.Id, orgId); err != nil {
			_ = tx.Rollback()
			c.Logger().Errorf("error del user org rel, err %s", err)
			return nil, err
		}
	}

	// 删除用户
	if err := services.DeleteUser(tx, user.Id); err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("error del user, err %s", err)
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("error commit del user, err %s", err)
		return nil, e.New(e.DBError, err)
	}

	return nil, nil
}
