package apps

import (
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
	"cloudiac/utils"
	"cloudiac/utils/mail"
	"fmt"
	"net/http"
)

type emailInviteUserData struct {
	*models.User
	InitPass     string // 初始化密码
	Inviter      string // 邀请人名称
	Organization string // 加入目标组织名称
	IsNewUser    bool   // 是否创建新用户
}

var (
	emailSubjectInviteUser = "用户邀请通知【CloudIaC】"
	emailBodyInviteUser    = "尊敬的 {{.Name}}：\n\n{{.Inviter}} 邀请您使用 CloudIaC 服务，您将加入 {{.Organization}} 组织。\n\n{{if .IsNewUser}}这是您的登录详细信息：\n\n登录名：\t{{.Email}}\n密码：\t{{.InitPass}}\n\n为了保障您的安全，请立即登陆您的账号并修改初始密码。{{else}}请使用 {{.Email}} 登陆您的账号使用 CloudIaC 服务。{{end}}"
)

// CreateOrganization 创建组织
func CreateOrganization(c *ctx.ServiceCtx, form *forms.CreateOrganizationForm) (*models.Organization, e.Error) {
	c.AddLogField("action", fmt.Sprintf("create org %s", form.Name))

	// 创建组织
	org, err := services.CreateOrganization(c.DB(), models.Organization{
		Name:        form.Name,
		CreatorId:   c.UserId,
		Description: form.Description,
	})
	if err != nil && err.Code() == e.OrganizationAlreadyExists {
		return nil, e.New(err.Code(), err, http.StatusBadRequest)
	} else if err != nil {
		c.Logger().Errorf("error creating org, err %s", err)
		return nil, e.AutoNew(err, e.DBError)
	}

	return org, nil
}

// SearchOrganization 组织查询
func SearchOrganization(c *ctx.ServiceCtx, form *forms.SearchOrganizationForm) (interface{}, e.Error) {
	query := services.QueryOrganization(c.DB())
	if c.IsSuperAdmin {
		if form.Status != "" {
			query = query.Where("iac_org.status = ?", form.Status)
		}
	} else {
		query = query.Where("iac_org.id in (?)", services.UserOrgIds(c.UserId))
		query = query.Where("iac_org.status = 'enable'")
	}

	if form.Q != "" {
		query = query.WhereLike("iac_org.name", form.Q)
	}

	// 默认按创建时间逆序排序
	if form.SortField() == "" {
		query = query.Order("iac_org.created_at DESC")
	}
	rs, err := getPage(query, form, &models.OrgDetailResp{})
	if err != nil {
		c.Logger().Errorf("error get page, err %s", err)
	}
	return rs, err
}

// UpdateOrganization 组织编辑
func UpdateOrganization(c *ctx.ServiceCtx, form *forms.UpdateOrganizationForm) (*models.Organization, e.Error) {
	c.AddLogField("action", fmt.Sprintf("update org %s", form.Id))
	query := c.DB()

	if !c.IsSuperAdmin && c.OrgId == "" {
		return nil, e.New(e.PermissionDeny, fmt.Errorf("super admin required"), http.StatusBadRequest)
	}

	attrs := models.Attrs{}
	if form.HasKey("name") {
		attrs["name"] = form.Name
	}

	if form.HasKey("description") {
		attrs["description"] = form.Description
	}

	if form.HasKey("runnerId") {
		attrs["runner_id"] = form.RunnerId
	}

	// 变更组织状态
	if form.HasKey("status") {
		if _, err := ChangeOrgStatus(c, &forms.DisableOrganizationForm{Id: form.Id, Status: form.Status}); err != nil {
			return nil, err
		}
	}

	if !c.IsSuperAdmin {
		query = query.Where("id in (?)", services.UserOrgIds(c.UserId))
	}
	org, err := services.UpdateOrganization(query, form.Id, attrs)
	if err != nil && err.Code() == e.OrganizationAliasDuplicate {
		return nil, e.New(err.Code(), err, http.StatusBadRequest)
	} else if err != nil {
		c.Logger().Errorf("error update org, err %s", err)
		return nil, err
	}
	return org, nil
}

//ChangeOrgStatus 修改组织启用/禁用状态
func ChangeOrgStatus(c *ctx.ServiceCtx, form *forms.DisableOrganizationForm) (*models.Organization, e.Error) {
	c.AddLogField("action", fmt.Sprintf("change org status %s", form.Id))
	if !c.IsSuperAdmin && c.OrgId == "" {
		return nil, e.New(e.PermissionDeny, fmt.Errorf("super admin required"), http.StatusBadRequest)
	}
	query := c.DB()

	if form.Status != models.OrgEnable && form.Status != models.OrgDisable {
		return nil, e.New(e.OrganizationInvalidStatus, http.StatusBadRequest)
	}

	org, err := services.GetOrganizationById(query, form.Id)
	if err != nil && err.Code() == e.OrganizationNotExists {
		return nil, e.New(err.Code(), err, http.StatusBadRequest)
	} else if err != nil {
		c.Logger().Errorf("error get org by id, err %s", err)
		return nil, err
	}

	if org.Status == form.Status {
		return org, nil
	}

	if !c.IsSuperAdmin {
		query = services.QueryWithOrgId(query, c.OrgId)
	}
	org, err = services.UpdateOrganization(query, form.Id, models.Attrs{"status": form.Status})
	if err != nil && err.Code() == e.OrganizationAliasDuplicate {
		return nil, e.New(err.Code(), err, http.StatusBadRequest)
	} else if err != nil {
		c.Logger().Errorf("error update org, err %s", err)
		return nil, e.New(e.DBError, err)
	}

	return org, nil
}

type organizationDetailResp struct {
	models.Organization
	Creator string `json:"creator" example:"超级管理员"`
}

// OrganizationDetail 组织信息详情
func OrganizationDetail(c *ctx.ServiceCtx, form forms.DetailOrganizationForm) (*organizationDetailResp, e.Error) {
	var (
		org  *models.Organization
		user *models.User
		err  e.Error
	)
	query := c.DB()
	if !c.IsSuperAdmin {
		query = query.Where("id in (?)", services.UserOrgIds(c.UserId))
		query = query.Where("status = 'enable'")
	}

	org, err = services.GetOrganizationById(query, form.Id)
	if err != nil && err.Code() == e.OrganizationNotExists {
		return nil, e.New(e.OrganizationNotExists, err, http.StatusNotFound)
	} else if err != nil {
		c.Logger().Errorf("error get org by id, err %s", err)
		return nil, e.New(e.DBError, err)
	}
	user, err = services.GetUserById(c.DB(), org.CreatorId)
	if err != nil && err.Code() == e.UserNotExists {
		// 报 500 错误，正常情况用户不应该找不到，除非被意外删除
		return nil, e.New(e.UserNotExists, err, http.StatusInternalServerError)
	} else if err != nil {
		c.Logger().Errorf("error get user by id, err %s", err)
		return nil, e.New(e.DBError, err)
	}

	var o = organizationDetailResp{
		Organization: *org,
		Creator:      user.Name,
	}

	return &o, nil
}

// DeleteOrganization 删除组织
func DeleteOrganization(c *ctx.ServiceCtx, form *forms.DeleteOrganizationForm) (org *models.Organization, err e.Error) {
	c.AddLogField("action", fmt.Sprintf("delete org %s", form.Id))
	c.Logger().Errorf("del id %s", form.Id)
	return nil, e.New(e.BadRequest, http.StatusNotImplemented)
}

// DeleteUserOrgRel 从组织移除用户
func DeleteUserOrgRel(c *ctx.ServiceCtx, form *forms.DeleteUserOrgRelForm) (interface{}, e.Error) {
	c.AddLogField("action", fmt.Sprintf("delete user %s for org %s", form.UserId, c.OrgId))
	query := c.DB()
	query = query.Where("status = 'enable'")
	if !c.IsSuperAdmin {
		userIds, _ := services.GetUserIdsByOrg(c.DB(), c.OrgId)
		query = query.Where(fmt.Sprintf("%s.id in (?)", models.User{}.TableName()), userIds)
	}

	user, err := services.GetUserById(query, form.UserId)
	if err != nil && err.Code() == e.UserNotExists {
		return nil, e.New(err.Code(), err, http.StatusBadRequest)
	} else if err != nil {
		c.Logger().Errorf("error get user by id, err %s", err)
		return nil, e.New(e.DBError, err)
	}

	if err := services.DeleteUserOrgRel(c.DB(), form.UserId, c.OrgId); err != nil {
		c.Logger().Errorf("error del user org rel, err %s", err)
		return nil, err
	}
	c.Logger().Infof("delete user ", form.UserId, " for org ", c.OrgId, " succeed")

	resp := models.UserWithRoleResp{
		User: user,
		Role: "",
	}
	return resp, nil
}

// AddUserOrgRel 添加用户到组织
func AddUserOrgRel(c *ctx.ServiceCtx, form *forms.AddUserOrgRelForm) (*models.UserWithRoleResp, e.Error) {
	c.AddLogField("action", fmt.Sprintf("add user %s to org %s", form.UserId, form.Id))
	var user *models.User
	query := c.DB()
	query = query.Where("status = 'enable'")
	if !c.IsSuperAdmin {
		userIds, _ := services.GetUserIdsByOrg(c.DB(), c.OrgId)
		query = query.Where(fmt.Sprintf("%s.id in (?)", models.User{}.TableName()), userIds)
	}

	if form.Role != consts.OrgRoleMember && form.Role != consts.OrgRoleAdmin {
		return nil, e.New(e.InvalidRoleName, http.StatusBadRequest)
	}
	user, err := services.GetUserById(query, form.UserId)
	if err != nil && err.Code() == e.UserNotExists {
		return nil, e.New(err.Code(), err, http.StatusBadRequest)
	} else if err != nil {
		c.Logger().Errorf("error get user by id, err %s", err)
		return nil, e.New(e.DBError, err)
	}

	_, err = services.CreateUserOrgRel(c.DB(), models.UserOrg{OrgId: form.Id, UserId: form.UserId, Role: form.Role})
	if err != nil && err.Code() == e.UserAlreadyExists {
		c.Logger().Errorf("error create user org rel, err %s", err)
		return nil, e.New(err.Code(), err, http.StatusBadRequest)
	} else if err != nil {
		c.Logger().Errorf("error add user org rel, err %s", err)
		return nil, e.New(e.DBError, err)
	}
	c.Logger().Infof("add user ", form.Id, " to org ", c.OrgId, " succeed")

	resp := models.UserWithRoleResp{
		User: user,
		Role: form.Role,
	}

	return &resp, nil
}

// UpdateUserOrgRel 更新用户组织角色
func UpdateUserOrgRel(c *ctx.ServiceCtx, form *forms.UpdateUserOrgRelForm) (*models.UserWithRoleResp, e.Error) {
	c.AddLogField("action", fmt.Sprintf("update user %s in org %s to role %s", form.UserId, c.OrgId, form.Role))

	query := c.DB()
	query = query.Where("status = 'enable'")
	if !c.IsSuperAdmin {
		userIds, _ := services.GetUserIdsByOrg(c.DB(), c.OrgId)
		query = query.Where(fmt.Sprintf("%s.id in (?)", models.User{}.TableName()), userIds)
	}
	user, err := services.GetUserById(query, form.UserId)
	if err != nil && err.Code() == e.UserNotExists {
		return nil, e.New(err.Code(), err, http.StatusBadRequest)
	} else if err != nil {
		c.Logger().Errorf("error get user by id, err %s", err)
		return nil, e.New(e.DBError, err)
	}

	if err := services.UpdateUserOrgRel(c.DB(), models.UserOrg{OrgId: c.OrgId, UserId: form.UserId, Role: form.Role}); err != nil {
		c.Logger().Errorf("error create user org rel, err %s", err)
		return nil, err
	}
	c.Logger().Infof("add user ", form.UserId, " to org ", c.OrgId, " succeed")

	resp := models.UserWithRoleResp{
		User: user,
		Role: form.Role,
	}

	return &resp, nil
}

// InviteUser 邀请用户加入某个组织
// 如果用户不存在，则创建并加入组织，如果用户已经存在，则加入该组织
func InviteUser(c *ctx.ServiceCtx, form *forms.InviteUserForm) (*models.UserWithRoleResp, e.Error) {
	c.AddLogField("action", fmt.Sprintf("invite user %s%s to org %s as %s", form.Name, form.UserId, form.Id, form.Role))

	org, err := services.GetOrganizationById(c.DB(), form.Id)
	if err != nil && err.Code() == e.OrganizationNotExists {
		return nil, e.New(e.BadRequest, http.StatusBadRequest)
	} else if err != nil {
		c.Logger().Errorf("error get org, err %s", err)
		return nil, e.New(e.DBError, err)
	}
	if form.Role == "" {
		form.Role = consts.OrgRoleMember
	}

	tx := c.Tx()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	// 检查用户是否存在
	var user *models.User

	if form.UserId != "" {
		user, err = services.GetUserById(tx, form.UserId)
		if err != nil && err.Code() == e.UserNotExists {
			return nil, e.New(err.Code(), err, http.StatusBadRequest)
		} else if err != nil {
			c.Logger().Errorf("error get user by id, err %s", err)
			return nil, err
		}
	} else if form.Email != "" {
		// 查找系统是否已经存在该邮箱对应的用户
		user, err = services.GetUserByEmail(tx, form.Email)
		if err != nil && err.Code() != e.UserNotExists {
			c.Logger().Errorf("error get user by email, err %s", err)
			return nil, err
		}
	} else if form.Name == "" || form.Email == "" {
		return nil, e.New(e.BadRequest, http.StatusBadRequest)
	}

	initPass := utils.GenPasswd(6, "mix")
	hashedPassword, err := services.HashPassword(initPass)
	if err != nil {
		c.Logger().Errorf("error hash password, err %s", err)
		return nil, err
	}
	isNew := false
	if user == nil {
		user, err = services.CreateUser(tx, models.User{
			Name:     form.Name,
			Password: hashedPassword,
			Email:    form.Email,
			Phone:    form.Phone,
		})
		if err != nil && err.Code() == e.UserAlreadyExists {
			return nil, e.New(err.Code(), err, http.StatusBadRequest)
		} else if err != nil {
			_ = tx.Rollback()
			c.Logger().Errorf("error create user, err %s", err)
			return nil, err
		}
		isNew = true
	}

	// 建立用户与组织间关联
	if !isNew {
		if err := services.DeleteUserOrgRel(tx, user.Id, form.Id); err != nil {
			_ = tx.Rollback()
			c.Logger().Errorf("error del user org rel, err %s", err)
		}
	}
	if _, err = services.CreateUserOrgRel(tx, models.UserOrg{
		OrgId:  form.Id,
		UserId: user.Id,
		Role:   form.Role,
	}); err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("error create user org rel, err %s", err)
		return nil, err
	}

	// 新用户自动加入演示组织
	if isNew {
		if err = services.TryAddDemoRelation(tx, user.Id); err != nil {
			_ = tx.Rollback()
			c.Logger().Errorf("error add user demo rel, err %s", err)
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("error commit invite user, err %s", err)
		return nil, e.New(e.DBError, err)
	}

	// 发送邀请邮件
	data := emailInviteUserData{
		User:         user,
		IsNewUser:    isNew,
		Inviter:      c.Username,
		Organization: org.Name,
		InitPass:     initPass,
	}
	go func() {
		err := mail.SendMail([]string{user.Email}, emailSubjectInviteUser, utils.SprintTemplate(emailBodyInviteUser, data))
		if err != nil {
			c.Logger().Errorf("error send mail to %s, err %s", user.Email, err)
		}
	}()

	resp := models.UserWithRoleResp{
		User: user,
		Role: form.Role,
	}

	return &resp, nil
}
