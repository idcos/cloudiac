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

// CreateOrganization 创建组织
func CreateOrganization(c *ctx.ServiceCtx, form *forms.CreateOrganizationForm) (*models.Organization, e.Error) {
	c.AddLogField("action", fmt.Sprintf("create org %s", form.Name))

	// 检查管理员用户是否存在
	var (
		owner *models.User
		err   e.Error
	)

	tx := c.Tx()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	if form.OwnerId != "" {
		owner, err = services.GetUserById(tx, form.OwnerId)
		if err != nil && err.Code() == e.UserNotExists {
			return nil, e.New(err.Code(), err, http.StatusBadRequest)
		} else if err != nil {
			c.Logger().Errorf("error get user by id, err %s", err)
			return nil, err
		}
	} else if form.OwnerEmail != "" {
		// 查找系统是否已经存在该邮箱对应的用户
		owner, err = services.GetUserByEmail(tx, form.OwnerEmail)
		if err != nil && err.Code() != e.UserNotExists {
			c.Logger().Errorf("error get user by id, err %s", err)
			return nil, err
		}
	} else if form.OwnerName == "" || form.OwnerEmail == "" {
		return nil, e.New(e.BadRequest, http.StatusBadRequest)
	}

	// 创建组织
	org, err := services.CreateOrganization(tx, models.Organization{
		Name:        form.Name,
		CreatorId:   c.UserId,
		Description: form.Description,
	})
	if err != nil && err.Code() == e.OrganizationAlreadyExists {
		return nil, e.New(err.Code(), err, http.StatusBadRequest)
	} else if err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("error creating org, err %s", err)
		return nil, e.AutoNew(err, e.DBError)
	}

	// 创建组织管理员
	isNew := false
	if owner == nil {
		initPass := utils.GenPasswd(6, "mix")
		hashedPassword, err := services.HashPassword(initPass)
		if err != nil {
			_ = tx.Rollback()
			c.Logger().Errorf("error hash password, err %s", err)
			return nil, err
		}
		owner, err = services.CreateUser(tx, models.User{
			Name:     form.Name,
			Password: hashedPassword,
			Email:    form.OwnerEmail,
		})
		if err != nil && err.Code() != e.UserAlreadyExists {
			_ = tx.Rollback()
			c.Logger().Errorf("error create user, err %s", err)
			return nil, err
		}
		isNew = true
	}

	// 设置管理员
	if _, err := services.CreateUserOrgRel(tx, models.UserOrg{
		OrgId:  org.Id,
		UserId: owner.Id,
		Role:   consts.OrgRoleOwner,
	}); err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("error creating user org rel, err %s", err)
		return nil, e.AutoNew(err, e.DBError)
	}

	if err := tx.Commit(); err != nil {
		c.Logger().Errorf("error commit create org, err %s", err)
		_ = tx.Rollback()
		return nil, e.New(e.DBError, err)
	}

	// 发送邀请邮件
	data := emailInviteUserData{
		User:         owner,
		Inviter:      c.Username,
		Organization: org.Name,
		IsNewUser:    isNew,
	}

	// 发送邀请邮件
	go func() {
		err := mail.SendMail([]string{form.OwnerEmail}, emailSubjectInviteUser, utils.SprintTemplate(emailBodyInviteUser, data))
		if err != nil {
			c.Logger().Errorf("error send mail to %s, err %s", owner.Email, err)
		}
	}()

	return org, nil
}

// SearchOrganization 组织查询
func SearchOrganization(c *ctx.ServiceCtx, form *forms.SearchOrganizationForm) (interface{}, e.Error) {
	query := services.QueryOrganization(c.DB())
	if c.IsSuperAdmin == true {
		if form.Status != "" {
			query = query.Where("status = ?", form.Status)
		}
	} else {
		query = query.Where("status = 'enable'")
		orgIds, er := services.GetOrgIdsByUser(c.DB(), c.UserId)
		if er != nil {
			c.Logger().Errorf("error get org id by user, err %s", er)
			return nil, e.New(e.DBError, er)
		}
		query = query.Where("id in (?)", orgIds)
	}

	if form.Q != "" {
		query = query.WhereLike("name", form.Q)
	}

	// 默认按创建时间逆序排序
	if form.SortField() == "" {
		query = query.Order("created_at DESC")
	}
	rs, err := getPage(query, form, &models.Organization{})
	if err != nil {
		c.Logger().Errorf("error get page, err %s", err)
	}
	return rs, err
}

// UpdateOrganization 组织编辑
func UpdateOrganization(c *ctx.ServiceCtx, form *forms.UpdateOrganizationForm) (*models.Organization, e.Error) {
	c.AddLogField("action", fmt.Sprintf("update org %s", form.Id))

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

	org, err := services.UpdateOrganization(c.DB(), form.Id, attrs)
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

	if form.Status != models.OrgEnable && form.Status != models.OrgDisable {
		return nil, e.New(e.OrganizationInvalidStatus, http.StatusBadRequest)
	}

	org, err := services.GetOrganizationById(c.DB(), form.Id)
	if err != nil && err.Code() == e.OrganizationNotExists {
		return nil, e.New(err.Code(), err, http.StatusBadRequest)
	} else if err != nil {
		c.Logger().Errorf("error get org by id, err %s", err)
		return nil, err
	}

	if org.Status == form.Status {
		return org, nil
	}

	org, err = services.UpdateOrganization(c.DB(), form.Id, models.Attrs{"status": form.Status})
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
	orgIds, er := services.GetOrgIdsByUser(c.DB(), c.UserId)
	if er != nil {
		c.Logger().Errorf("error get org id by user, err %s", er)
		return nil, e.New(e.DBError, er)
	}
	if form.Id.InArray(orgIds...) == false && c.IsSuperAdmin == false {
		// 请求了一个不存在的 org，因为 org id 是在 path 传入，这里我们返回 404
		return nil, e.New(e.OrganizationNotExists, http.StatusNotFound)
	}

	var (
		org  *models.Organization
		user *models.User
		err  e.Error
	)
	org, err = services.GetOrganizationById(c.DB(), form.Id)
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

	tx := c.Tx()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := services.DeleteUserOrgRel(tx, form.UserId, c.OrgId); err != nil {
		tx.Rollback()
		c.Logger().Errorf("error del user org rel, err %s", err)
		return nil, err
	} else if err := tx.Commit(); err != nil {
		tx.Rollback()
		c.Logger().Errorf("error commit del user org rel, err %s", err)
		return nil, e.New(e.DBError, err)
	}
	c.Logger().Infof("delete user ", form.UserId, " for org ", c.OrgId, " succeed")

	user, _ := services.GetUserById(tx, form.UserId)
	return user, nil
}

// AddUserOrgRel 添加用户到组织
func AddUserOrgRel(c *ctx.ServiceCtx, form *forms.AddUserOrgRelForm) (*UserWithRoleResp, e.Error) {
	c.AddLogField("action", fmt.Sprintf("add user %s to org %s", form.Id, c.OrgId))

	tx := c.Tx()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if form.Role != consts.OrgRoleMember && form.Role != consts.OrgRoleOwner {
		return nil, e.New(e.InvalidRoleName, http.StatusBadRequest)
	}

	_, err := services.CreateUserOrgRel(tx, models.UserOrg{OrgId: c.OrgId, UserId: form.Id, Role: form.Role})
	if err != nil && err.Code() != e.UserAlreadyExists {
		tx.Rollback()
		c.Logger().Errorf("error create user org rel, err %s", err)
		return nil, err
	} else if err := tx.Commit(); err != nil {
		tx.Rollback()
		c.Logger().Errorf("error commit add user org rel, err %s", err)
		return nil, e.New(e.DBError, err)
	}
	c.Logger().Infof("add user ", form.Id, " to org ", c.OrgId, " succeed")

	user, _ := services.GetUserById(tx, form.Id)
	resp := UserWithRoleResp{
		User: *user,
		Role: form.Role,
	}

	return &resp, nil
}

// UpdateUserOrgRel 更新用户组织角色
func UpdateUserOrgRel(c *ctx.ServiceCtx, form *forms.UpdateUserOrgRelForm) (*UserWithRoleResp, e.Error) {
	c.AddLogField("action", fmt.Sprintf("update user %s in org %s to role %s", form.UserId, c.OrgId, form.Role))

	tx := c.Tx()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := services.UpdateUserOrgRel(tx, models.UserOrg{OrgId: c.OrgId, UserId: form.UserId, Role: form.Role}); err != nil {
		tx.Rollback()
		c.Logger().Errorf("error create user org rel, err %s", err)
		return nil, err
	} else if err := tx.Commit(); err != nil {
		tx.Rollback()
		c.Logger().Errorf("error commit add user org rel, err %s", err)
		return nil, e.New(e.DBError, err)
	}
	c.Logger().Infof("add user ", form.UserId, " to org ", c.OrgId, " succeed")

	user, _ := services.GetUserById(tx, form.UserId)
	resp := UserWithRoleResp{
		User: *user,
		Role: form.Role,
	}

	return &resp, nil
}
