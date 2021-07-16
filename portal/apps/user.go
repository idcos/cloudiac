package apps

import (
	"cloudiac/portal/consts"
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
	InitPass string `json:"initPass,omitempty" example:"rANd0m"` // 初始化密码
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
		User: user,
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

type UserWithRoleResp struct {
	*models.User
	Password    string `json:"-"`
	Role        string `json:"role,omitempty" example:"member"`         // 组织角色
	ProjectRole string `json:"projectRole,omitempty" example:"manager"` // 项目角色
}

// SearchUser 查询用户列表
func SearchUser(c *ctx.ServiceCtx, form *forms.SearchUserForm) (interface{}, e.Error) {
	query := c.DB()
	// 管理员查看所有用户
	if c.OrgId == "" && c.ProjectId == "" && !c.IsSuperAdmin {
		return nil, e.New(e.PermissionDeny, fmt.Errorf("super admin required"), http.StatusBadRequest)
	}
	// 组织成员查询
	if c.OrgId != "" && c.ProjectId == "" {
		query = services.QueryWithOrgId(query, c.OrgId)
	}
	// 项目成员查询
	if c.OrgId != "" && c.ProjectId != "" {
		query = services.QueryWithProjectId(query, c.ProjectId)
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
	users := make([]*UserWithRoleResp, 0)
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
func UpdateUser(c *ctx.ServiceCtx, form *forms.UpdateUserForm) (user *models.User, err e.Error) {
	c.AddLogField("action", fmt.Sprintf("update user %s", form.Id))
	if c.IsSuperAdmin == false && c.Role != consts.OrgRoleAdmin && c.UserId != form.Id {
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
	if c.IsSuperAdmin == false {
		return nil, e.New(e.PermissionDeny, http.StatusForbidden)
	}

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
	query := c.DB()
	// 组织用户查询
	if c.OrgId != "" {
		query = services.QueryWithOrgId(query, c.OrgId)
	}
	// 项目用户查询
	if c.ProjectId != "" {
		query = services.QueryWithProjectId(query, c.ProjectId)
	}
	if c.UserId == id {
		// 查自己
	} else if (c.OrgId == "" && c.ProjectId == "") && !c.IsSuperAdmin {
		// 全局用户查询，需要管理员权限
		return nil, e.New(e.PermissionDeny, http.StatusForbidden)
	}

	user, err := services.GetUserById(query, id)
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
