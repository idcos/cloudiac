package apps

import (
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/libs/db"
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
	InitPass string `json:"initPass,omitempty" example:"rANd0m"` // 初始化密码
}

var (
	emailSubjectCreateUser = "注册用户成功通知"
	emailBodyCreateUser    = "尊敬的 {{.Name}}：\n\n恭喜您完成注册 CloudIaC 服务。\n\n这是您的登录详细信息：\n\n登录名：\t{{.Email}}\n密码：\t{{.InitPass}}\n\n为了保障您的安全，请立即登陆您的账号并修改初始密码。"
)

var (
	emailSubjectResetPassword = "密码重置通知【CloudIaC】"
	emailBodyResetPassword    = "尊敬的 {{.Name}}：\n\n您的密码已经被重置，这是您的新密码：\n\n密码：\t{{.InitPass}}\n\n请使用新密码登陆系统。\n\n为了保障您的安全，请立即登陆您的账号并修改密码。"
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

	tx := c.Tx()
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

	// 发送邮件给用户
	go func() {
		err := mail.SendMail([]string{user.Email}, emailSubjectCreateUser, utils.SprintTemplate(emailBodyCreateUser, resp))
		if err != nil {
			c.Logger().Errorf("error send mail to %s, err %s", user.Email, err)
		}
	}()

	return &resp, nil
}

// SearchUser 查询用户列表
func SearchUser(c *ctx.ServiceCtx, form *forms.SearchUserForm) (interface{}, e.Error) {
	query := services.QueryUser(c.DB())

	if c.OrgId == "" && !c.IsSuperAdmin {
		return nil, e.New(e.PermissionDeny, fmt.Errorf("super admin required"), http.StatusBadRequest)
	}

	if c.OrgId != "" {
		userIds, _ := services.GetUserIdsByOrg(c.DB(), c.OrgId)
		query = query.Where(fmt.Sprintf("%s.id in (?)", models.User{}.TableName()), userIds)
	}
	if c.ProjectId != "" {
		userIds, _ := services.GetUserIdsByProject(c.DB(), c.ProjectId)
		query = query.Where(fmt.Sprintf("%s.id in (?)", models.User{}.TableName()), userIds)
	}

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

	// 导出用户角色
	if c.OrgId != "" {
		query = query.Joins(fmt.Sprintf("left join %s as o on %s.id = o.user_id and o.org_id = ?",
			models.UserOrg{}.TableName(), models.User{}.TableName()), c.OrgId).
			LazySelectAppend(fmt.Sprintf("o.role,%s.*", models.User{}.TableName()))
	}
	// 导出项目角色
	if c.ProjectId != "" {
		query = query.Joins(fmt.Sprintf("left join %s as p on %s.id = p.user_id and p.project_id = ?",
			models.UserProject{}.TableName(), models.User{}.TableName()), c.ProjectId).
			LazySelectAppend(fmt.Sprintf("p.role as project_role,%s.*", models.User{}.TableName()))
	}

	p := page.New(form.CurrentPage(), form.PageSize(), query)
	users := make([]*models.UserWithRoleResp, 0)
	if err := p.Scan(&users); err != nil {
		c.Logger().Errorf("error get users, err %s", err)
		return nil, e.New(e.DBError, err)
	}

	return page.PageResp{
		Total:    p.MustTotal(),
		PageSize: p.Size,
		List:     users,
	}, nil
}

// UpdateUser 用户信息编辑
func UpdateUser(c *ctx.ServiceCtx, form *forms.UpdateUserForm) (*models.User, e.Error) {
	c.AddLogField("action", fmt.Sprintf("update user %s", form.Id))
	if c.UserId == form.Id {
		// 自身编辑
	} else if c.OrgId != "" && !services.UserHasOrgRole(c.UserId, c.OrgId, consts.OrgRoleAdmin) {
		return nil, e.New(e.PermissionDeny, fmt.Errorf("admin required"), http.StatusForbidden)
	} else if c.OrgId == "" && !c.IsSuperAdmin {
		return nil, e.New(e.PermissionDeny, fmt.Errorf("super admin required"), http.StatusForbidden)
	}

	query := c.DB()
	user, err := services.GetUserById(query, form.Id)
	if err != nil {
		return nil, e.New(err.Code(), err, http.StatusBadRequest)
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

	return services.UpdateUser(c.DB(), form.Id, attrs)
}

// ChangeUserStatus 修改用户启用/禁用状态
// 需要平台管理员权限。
func ChangeUserStatus(c *ctx.ServiceCtx, form *forms.DisableUserForm) (*models.User, e.Error) {
	c.AddLogField("action", fmt.Sprintf("change user status %s", form.Id))
	query := c.DB()
	if c.IsSuperAdmin == false {
		return nil, e.New(e.PermissionDeny, http.StatusForbidden)
	}

	if form.Status != models.Enable && form.Status != models.Disable {
		return nil, e.New(e.UserInvalidStatus, http.StatusBadRequest)
	}

	user, err := services.GetUserById(query, form.Id)
	if err != nil && err.Code() == e.UserNotExists {
		return nil, e.New(err.Code(), err, http.StatusBadRequest)
	} else if err != nil {
		c.Logger().Errorf("error get user by id, err %s", err)
		return nil, err
	}

	if user.Status == form.Status {
		return user, nil
	}

	user, err = services.UpdateUser(query, form.Id, models.Attrs{"status": form.Status})
	if err != nil {
		c.Logger().Errorf("error update user, err %s", err)
		return nil, e.New(e.DBError, err)
	}

	return user, nil
}

// UserDetail 获取单个用户详情
func UserDetail(c *ctx.ServiceCtx, userId models.Id) (*models.UserWithRoleResp, e.Error) {
	query := c.DB()

	if c.OrgId != "" {
		userIds, _ := services.GetUserIdsByOrg(c.DB(), c.OrgId)
		query = query.Where(fmt.Sprintf("%s.id in (?)", models.User{}.TableName()), userIds)
	}
	if c.ProjectId != "" {
		userIds, _ := services.GetUserIdsByProject(c.DB(), c.ProjectId)
		query = query.Where(fmt.Sprintf("%s.id in (?)", models.User{}.TableName()), userIds)
	}
	if c.UserId == userId {
		// 查自己
	} else if (c.OrgId == "" && c.ProjectId == "") && !c.IsSuperAdmin {
		// 全局用户查询，需要管理员权限
		return nil, e.New(e.PermissionDeny, fmt.Errorf("super admin required"), http.StatusForbidden)
	}

	// 导出用户角色
	if c.OrgId != "" {
		query = query.Joins(fmt.Sprintf("left join %s as o on %s.id = o.user_id and o.org_id = ?",
			models.UserOrg{}.TableName(), models.User{}.TableName()), c.OrgId).
			LazySelectAppend(fmt.Sprintf("o.role,%s.*", models.User{}.TableName()))
	}
	// 导出项目角色
	if c.ProjectId != "" {
		query = query.Joins(fmt.Sprintf("left join %s as p on %s.id = p.user_id and p.project_id = ?",
			models.UserProject{}.TableName(), models.User{}.TableName()), c.ProjectId).
			LazySelectAppend(fmt.Sprintf("p.role as project_role,%s.*", models.User{}.TableName()))
	}
	detail, err := services.GetUserDetailById(query, userId)
	if err != nil && err.Code() == e.UserNotExists {
		// 通过 /auth/me 或者 /users/:userId 访问
		return nil, e.New(err.Code(), err, http.StatusNotFound)
	} else if err != nil {
		c.Logger().Errorf("error get user by id, err %s", err)
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	}

	return detail, nil
}

// DeleteUser 删除用户
// 需要平台管理员权限。
func DeleteUser(c *ctx.ServiceCtx, form *forms.DeleteUserForm) (interface{}, e.Error) {
	c.AddLogField("action", fmt.Sprintf("delete user %s", form.Id))
	if !c.IsSuperAdmin {
		return nil, e.New(e.PermissionDeny, http.StatusForbidden)
	}

	tx := c.Tx()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	user, err := services.GetUserById(tx, form.Id)
	c.Logger().Errorf("del user %s", form.Id)
	if err != nil && err.Code() == e.UserNotExists {
		return nil, e.New(err.Code(), err, http.StatusNotFound)
	} else if err != nil {
		c.Logger().Errorf("error get user by id, err %s", err)
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	}

	// 解除组织关系
	orgIds, er := services.GetOrgIdsByUser(tx, c.UserId)
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

//
//// UserRestrictOrg 获取用户访问范围限制
//func UserRestrictOrg(c *ctx.ServiceCtx, query *db.Session) *db.Session {
//	query = query.Model(models.User{})
//	if c.OrgId != "" {
//		subQ := query.Model(models.UserOrg{}).Select("user_id").Where("org_id = ?", c.OrgId)
//		query = query.Where(fmt.Sprintf("%s.id in (?)", models.User{}.TableName()), subQ.Expr())
//	} else {
//		// 如果是管理员，不需要附加限制参数，返回所有数据
//		// 组织管理员或者普通用户，如果不带 org，应该返回该用户关联的所有 org
//		if !c.IsSuperAdmin {
//			//select DISTINCT user_id from iac_user_org where (org_id in
//			//   (SELECT org_id from iac_user_org WHERE user_id = 'u-c3i41c06n88g4a2pet20'))
//			orgQ := query.Model(models.UserOrg{}).Select("org_id").Where("user_id = ?", c.UserId)
//			subQ := query.Model(models.UserOrg{}).Select("DISTINCT user_id").Where("org_id in (?)", orgQ)
//			query = query.Where(fmt.Sprintf("%s.id in (?)", models.User{}.TableName()), subQ.Expr())
//		}
//	}
//	return query
//}

func QueryUserWithUserIdsByOrg(query *db.Session, orgId models.Id) *db.Session {
	userIds, _ := services.GetUserIdsByOrg(query, orgId)
	query = query.Where(fmt.Sprintf("%s.id in (?)", models.User{}.TableName()), userIds)
	return query
}

func QueryUserWithUserIdsByProject(query *db.Session, projectId models.Id) *db.Session {
	userIds, _ := services.GetUserIdsByProject(query, projectId)
	query = query.Where(fmt.Sprintf("%s.id in (?)", models.User{}.TableName()), userIds)
	return query
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

	// 发送密码重置通知邮件
	go func() {
		err := mail.SendMail([]string{user.Email}, emailSubjectResetPassword, utils.SprintTemplate(emailBodyResetPassword, resp))
		if err != nil {
			c.Logger().Errorf("error send mail to %s, err %s", user.Email, err)
		}
	}()

	return user, err
}
