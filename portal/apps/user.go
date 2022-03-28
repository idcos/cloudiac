// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package apps

import (
	"cloudiac/common"
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/libs/page"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/models/resps"
	"cloudiac/portal/services"
	"cloudiac/utils"
	"cloudiac/utils/logs"
	"cloudiac/utils/mail"
	"encoding/json"
	"fmt"
	"net/http"
)

// CreateUserResp 创建用户返回结果，带上初始化的随机密码

var (
	emailSubjectCreateUser = "注册用户成功通知"
	emailBodyCreateUser    = "尊敬的 {{.Name}}：\n\n恭喜您完成注册 CloudIaC 服务。\n\n这是您的登录详细信息：\n\n登录名：\t{{.Email}}\n密码：\t{{.InitPass}}\n\n为了保障您的安全，请立即登陆您的账号并修改初始密码。"
)

var (
	emailSubjectResetPassword = "密码重置通知【CloudIaC】"                                                                                      //nolint:gosec
	emailBodyResetPassword    = "尊敬的 {{.Name}}：\n\n您的密码已经被重置，这是您的新密码：\n\n密码：\t{{.InitPass}}\n\n请使用新密码登陆系统。\n\n为了保障您的安全，请立即登陆您的账号并修改密码。" //nolint:gosec
)

func createUserOrgRel(tx *db.Session, orgId models.Id, initPass string, form *forms.CreateUserForm, lg logs.Logger) (*models.User, e.Error) {
	var (
		user *models.User
		err  e.Error
	)

	hashedPassword, err := services.HashPassword(initPass)
	if err != nil {
		lg.Errorf("error hash password, err %s", err)
		return nil, err
	}

	user, err = services.CreateUser(tx, models.User{
		Name:     form.Name,
		Password: hashedPassword,
		Phone:    form.Phone,
		Email:    form.Email,
	})
	if err != nil && err.Code() == e.UserAlreadyExists {
		return nil, e.New(err.Code(), err, http.StatusBadRequest)
	} else if err != nil {
		lg.Errorf("error create user, err %s", err)
		return nil, err
	}

	// 建立用户与组织间关联
	_, err = services.CreateUserOrgRel(tx, models.UserOrg{
		OrgId:  orgId,
		UserId: user.Id,
	})
	if err != nil {
		lg.Errorf("error create user , err %s", err)
		return nil, err
	}

	// 新用户自动加入演示组织和项目
	if orgId != models.Id(common.DemoOrgId) {
		if err = services.TryAddDemoRelation(tx, user.Id); err != nil {
			_ = tx.Rollback()
			lg.Errorf("error add user demo rel, err %s", err)
			return nil, err
		}
	}

	return user, nil
}

// CreateUser 创建用户
func CreateUser(c *ctx.ServiceContext, form *forms.CreateUserForm) (*resps.CreateUserResp, e.Error) {
	c.AddLogField("action", fmt.Sprintf("create user %s", form.Name))

	tx := c.Tx()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	initPass := utils.GenPasswd(6, "mix")
	user, err := createUserOrgRel(tx, c.OrgId, initPass, form, c.Logger())
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
	resp := resps.CreateUserResp{
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

func queryUserOrg(db, query *db.Session, orgId models.Id, isSuperAdmin bool, exclude string) (*db.Session, e.Error) {
	if orgId == "" && !isSuperAdmin {
		return nil, e.New(e.PermissionDeny, fmt.Errorf("super admin required"), http.StatusBadRequest)
	}

	if orgId == "" {
		return query, nil
	}

	userIds, _ := services.GetUserIdsByOrg(db, orgId)
	if exclude == "org" {
		// 排除组织已有用户
		// 应只有平台管理员可以调用
		if !isSuperAdmin {
			return nil, e.New(e.PermissionDeny, fmt.Errorf("super admin required"), http.StatusBadRequest)
		}
		rootIds, _ := services.GetRootUserIds(db)
		userIds = append(userIds, rootIds...)
		return query.Where(fmt.Sprintf("%s.id not in (?)", models.User{}.TableName()), userIds), nil
	} else {
		// 查询组织所有用户
		return query.Where(fmt.Sprintf("%s.id in (?)", models.User{}.TableName()), userIds), nil
	}
}

func queryUserProject(db, query *db.Session, orgId, projectId models.Id, exclude string) (*db.Session, e.Error) {

	if projectId == "" {
		return query, nil
	}

	userIds, _ := services.GetUserIdsByProject(db, projectId)
	if exclude == "project" {
		// 排除组织里面的项目用户，包括所有组织管理员（自动获得项目权限）和已经加入组织的用户
		orgUserIds, _ := services.GetUserIdsByOrg(db, orgId)

		orgAdminsIds, _ := services.GetOrgAdminsByOrg(db, orgId)
		rootIds, _ := services.GetRootUserIds(db)
		excludeIds := append(userIds, orgAdminsIds...)
		excludeIds = append(excludeIds, rootIds...)

		query = query.Where(fmt.Sprintf("%s.id in (?)", models.User{}.TableName()), orgUserIds)
		return query.Where(fmt.Sprintf("%s.id not in (?)", models.User{}.TableName()), excludeIds), nil
	} else {
		return query.Where(fmt.Sprintf("%s.id in (?)", models.User{}.TableName()), userIds), nil
	}
}

// 提供私有化部署的用户搜索接口
func SearchAllUser(c *ctx.ServiceContext, form *forms.SearchUserForm) (interface{}, e.Error) {
	query := services.QueryUser(c.DB())
	return doSearchUser(c, query, form, true)
}

// SearchUser 查询用户列表
func SearchUser(c *ctx.ServiceContext, form *forms.SearchUserForm) (interface{}, e.Error) {
	query := services.QueryUser(c.DB())
	query, err := queryUserOrg(c.DB(), query, c.OrgId, c.IsSuperAdmin, form.Exclude)
	if err != nil {
		return nil, err
	}
	query, err = queryUserProject(c.DB(), query, c.OrgId, c.ProjectId, form.Exclude)
	if err != nil {
		return nil, err
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
	return doSearchUser(c, query, form, false)
}

func doSearchUser(c *ctx.ServiceContext, query *db.Session, form *forms.SearchUserForm, isLimit bool) (interface{}, e.Error) {
	var (
		currentPage int
		limit       int
	)
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
	if isLimit {
		limit = 10
		currentPage = 1
	} else {
		limit = form.PageSize()
		currentPage = form.CurrentPage()
	}
	p := page.New(currentPage, limit, query)
	users := make([]*resps.UserWithRoleResp, 0)
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

func chkUserIdentity(formId, userId, orgId models.Id, isSuperAdmin bool) e.Error {

	if formId == consts.SysUserId {
		return e.New(e.PermissionDeny, fmt.Errorf("modify sys user denied"), http.StatusForbidden)
	} else if userId == formId || isSuperAdmin {
		// 自身编辑
	} else if orgId != "" && !services.UserHasOrgRole(userId, orgId, consts.OrgRoleAdmin) {
		return e.New(e.PermissionDeny, fmt.Errorf("admin required"), http.StatusForbidden)
	} else if orgId == "" && !isSuperAdmin {
		return e.New(e.PermissionDeny, fmt.Errorf("super admin required"), http.StatusForbidden)
	}

	return nil
}

func getNewPassword(oldPassword, newPassword, userPassword string) (string, e.Error) {

	valid, err := utils.CheckPassword(oldPassword, userPassword)
	if err != nil {
		return "", e.New(e.DBError, http.StatusInternalServerError, err)
	}
	if !valid {
		return "", e.New(e.InvalidPassword, http.StatusBadRequest)
	}

	newPassword, er := services.HashPassword(newPassword)
	if er != nil {
		return "", er
	}
	return newPassword, nil
}

// UpdateUser 用户信息编辑
func UpdateUser(c *ctx.ServiceContext, form *forms.UpdateUserForm) (*models.User, e.Error) {
	c.AddLogField("action", fmt.Sprintf("update user %s", form.Id))
	err := chkUserIdentity(form.Id, c.UserId, c.OrgId, c.IsSuperAdmin)
	if err != nil {
		return nil, err
	}

	query := c.DB()
	user, err := services.GetUserById(query, form.Id)
	if err != nil {
		return nil, e.New(err.Code(), err, http.StatusBadRequest)
	}
	if user.IsLdap {
		return nil, e.New(e.LdapUpdateFailed)
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

	if !form.HasKey("oldPassword") {
		return services.UpdateUser(c.DB(), form.Id, attrs)
	}

	if !form.HasKey("newPassword") {
		return nil, e.New(e.BadParam, http.StatusBadRequest)
	}

	newPassword, er := getNewPassword(form.OldPassword, form.NewPassword, user.Password)
	if er != nil {
		return nil, er
	}
	attrs["password"] = newPassword

	return services.UpdateUser(c.DB(), form.Id, attrs)
}

// ChangeUserStatus 修改用户启用/禁用状态
// 需要平台管理员权限。
func ChangeUserStatus(c *ctx.ServiceContext, form *forms.DisableUserForm) (*models.User, e.Error) {
	c.AddLogField("action", fmt.Sprintf("change user status %s", form.Id))
	query := c.DB()
	if form.Id == consts.SysUserId {
		return nil, e.New(e.PermissionDeny, fmt.Errorf("modify sys user denied"), http.StatusForbidden)
	} else if !c.IsSuperAdmin {
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

func queryByOrgAndProject(db, query *db.Session, userId, orgId, projectId, inputUserId models.Id, isSuperAdmin bool) (*db.Session, e.Error) {
	if isSuperAdmin || userId == inputUserId {
		// 管理员查询任意用户或自身查询
	} else if orgId != "" {
		if projectId != "" {
			// 查询项目用户：组织管理员或项目成员
			if services.UserHasOrgRole(userId, orgId, consts.OrgRoleAdmin) ||
				services.UserHasProjectRole(userId, orgId, projectId, "") {
				userIds, _ := services.GetUserIdsByProject(db, projectId)
				return query.Where(fmt.Sprintf("%s.id in (?)", models.User{}.TableName()), userIds), nil
			} else {
				return nil, e.New(e.PermissionDeny, fmt.Errorf("project permission required"), http.StatusForbidden)
			}
		} else {
			// 查询组织用户
			if services.UserHasOrgRole(userId, orgId, "") {
				userIds, _ := services.GetUserIdsByOrg(db, orgId)
				return query.Where(fmt.Sprintf("%s.id in (?)", models.User{}.TableName()), userIds), nil
			} else {
				return nil, e.New(e.PermissionDeny, fmt.Errorf("org permission required"), http.StatusForbidden)
			}
		}
	} else {
		return nil, e.New(e.PermissionDeny, fmt.Errorf("super admin required"), http.StatusForbidden)
	}

	return query, nil
}

func setUserRole(detail *resps.UserWithRoleResp, userId, orgId, projectId models.Id, isSuperAdmin bool) {
	if isSuperAdmin {
		// 如果是平台管理员，自动拥有组织管理员权限和项目管理者权限
		if orgId != "" {
			detail.Role = consts.OrgRoleAdmin
		}
		if projectId != "" {
			detail.ProjectRole = consts.ProjectRoleManager
		}
	} else if services.UserHasOrgRole(userId, orgId, consts.OrgRoleAdmin) {
		// 如果是组织管理员自动拥有项目管理者权限
		if projectId != "" {
			detail.ProjectRole = consts.ProjectRoleManager
		}
	}
}

// UserDetail 获取单个用户详情
func UserDetail(c *ctx.ServiceContext, userId models.Id) (*resps.UserWithRoleResp, e.Error) {
	query := c.DB()
	query, err := queryByOrgAndProject(c.DB(), query, c.UserId, c.OrgId, c.ProjectId, userId, c.IsSuperAdmin)
	if err != nil {
		return nil, err
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

	setUserRole(detail, c.UserId, c.OrgId, c.ProjectId, c.IsSuperAdmin)
	return detail, nil
}

// DeleteUser 删除用户
// 需要平台管理员权限。
func DeleteUser(c *ctx.ServiceContext, form *forms.DeleteUserForm) (interface{}, e.Error) {
	c.AddLogField("action", fmt.Sprintf("delete user %s", form.Id))
	if form.Id == consts.SysUserId {
		return nil, e.New(e.PermissionDeny, fmt.Errorf("delete sys user denied"), http.StatusForbidden)
	} else if !c.IsSuperAdmin {
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
func UserPassReset(c *ctx.ServiceContext, form *forms.DetailUserForm) (*models.User, e.Error) {
	if form.Id == consts.SysUserId {
		return nil, e.New(e.PermissionDeny, fmt.Errorf("modify sys user denied"), http.StatusForbidden)
	}

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
