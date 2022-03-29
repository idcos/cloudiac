// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package apps

import (
	"cloudiac/common"
	"cloudiac/configs"
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
	"fmt"
	"net/http"
	"path"
	"strings"
)

type emailInviteUserData struct {
	*models.User
	InitPass     string // 初始化密码
	Inviter      string // 邀请人名称
	Organization string // 加入目标组织名称
	IsNewUser    bool   // 是否创建新用户
	Addr         string
}

var (
	emailSubjectInviteUser = "用户邀请通知【CloudIaC】"
)

// CreateOrganization 创建组织
func CreateOrganization(c *ctx.ServiceContext, form *forms.CreateOrganizationForm) (*models.Organization, e.Error) {
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
func SearchOrganization(c *ctx.ServiceContext, form *forms.SearchOrganizationForm) (interface{}, e.Error) {
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
	rs, err := getPage(query, form, &resps.OrgDetailResp{})
	if err != nil {
		c.Logger().Errorf("error get page, err %s", err)
	}
	return rs, err
}

// UpdateOrganization 组织编辑
func UpdateOrganization(c *ctx.ServiceContext, form *forms.UpdateOrganizationForm) (*models.Organization, e.Error) {
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
	if err != nil && err.Code() == e.OrganizationAlreadyExists {
		return nil, e.New(err.Code(), err, http.StatusBadRequest)
	} else if err != nil {
		c.Logger().Errorf("error update org, err %s", err)
		return nil, err
	}
	return org, nil
}

//ChangeOrgStatus 修改组织启用/禁用状态
func ChangeOrgStatus(c *ctx.ServiceContext, form *forms.DisableOrganizationForm) (*models.Organization, e.Error) {
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
	if err != nil && err.Code() == e.OrganizationAlreadyExists {
		return nil, e.New(err.Code(), err, http.StatusBadRequest)
	} else if err != nil {
		c.Logger().Errorf("error update org, err %s", err)
		return nil, e.New(e.DBError, err)
	}

	return org, nil
}

// OrganizationDetail 组织信息详情
func OrganizationDetail(c *ctx.ServiceContext, form forms.DetailOrganizationForm) (*resps.OrganizationDetailResp, e.Error) {
	var (
		org  *models.Organization
		user *models.User
		err  e.Error
	)
	query := c.DB()
	if !c.IsSuperAdmin {
		query = query.Where("id in (?) AND status = 'enable'", services.UserOrgIds(c.UserId))
	}

	org, err = services.GetOrganizationById(query, form.Id)
	if err != nil && err.Code() == e.OrganizationNotExists {
		return nil, e.New(e.OrganizationNotExists, err, http.StatusNotFound)
	} else if err != nil {
		c.Logger().Errorf("error get org by id, err %s", err)
		return nil, e.New(e.DBError, err)
	}
	user, err = services.GetUserByIdRaw(c.DB(), org.CreatorId)
	if err != nil && err.Code() == e.UserNotExists {
		// 报 500 错误，正常情况用户不应该找不到，除非被意外删除
		return nil, e.New(e.UserNotExists, err, http.StatusInternalServerError)
	} else if err != nil {
		c.Logger().Errorf("error get user by id, err %s", err)
		return nil, e.New(e.DBError, err)
	}

	var o = resps.OrganizationDetailResp{
		Organization: *org,
		Creator:      user.Name,
	}

	return &o, nil
}

// DeleteOrganization 删除组织
func DeleteOrganization(c *ctx.ServiceContext, form *forms.DeleteOrganizationForm) (org *models.Organization, err e.Error) {
	c.AddLogField("action", fmt.Sprintf("delete org %s", form.Id))
	c.Logger().Errorf("del id %s", form.Id)
	return nil, e.New(e.BadRequest, http.StatusNotImplemented)
}

// DeleteUserOrgRel 从组织移除用户
func DeleteUserOrgRel(c *ctx.ServiceContext, form *forms.DeleteUserOrgRelForm) (interface{}, e.Error) {
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
	if err := services.DeleteUserAllProject(c.DB(), form.UserId, c.OrgId); err != nil {
		c.Logger().Errorf("error del user project rel, err %s", err)
		return nil, err
	}

	c.Logger().Infof("delete user ", form.UserId, " for org ", c.OrgId, " succeed")

	resp := resps.UserWithRoleResp{
		User: *user,
		Role: "",
	}
	return resp, nil
}

// AddUserOrgRel 添加用户到组织
func AddUserOrgRel(c *ctx.ServiceContext, form *forms.AddUserOrgRelForm) (*resps.UserWithRoleResp, e.Error) {
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

	resp := resps.UserWithRoleResp{
		User: *user,
		Role: form.Role,
	}

	return &resp, nil
}

// UpdateUserOrgRel 更新用户组织角色
func UpdateUserOrgRel(c *ctx.ServiceContext, form *forms.UpdateUserOrgRelForm) (*resps.UserWithRoleResp, e.Error) {
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

	resp := resps.UserWithRoleResp{
		User: *user,
		Role: form.Role,
	}

	return &resp, nil
}

func getInviteUserOrg(c *ctx.ServiceContext, form *forms.InviteUserForm) (*models.Organization, e.Error) {
	org, err := services.GetOrganizationById(c.DB(), form.Id)
	if err != nil && err.Code() == e.OrganizationNotExists {
		return nil, e.New(e.BadRequest, http.StatusBadRequest)
	} else if err != nil {
		c.Logger().Errorf("error get org, err %s", err)
		return nil, e.New(e.DBError, err)
	}

	return org, nil
}

func getInviteUserById(c *ctx.ServiceContext, tx *db.Session, userId models.Id) (*models.User, e.Error) {
	user, err := services.GetUserById(tx, userId)
	if err != nil && err.Code() == e.UserNotExists {
		return nil, e.New(err.Code(), err, http.StatusBadRequest)
	} else if err != nil {
		c.Logger().Errorf("error get user by id, err %s", err)
		return nil, err
	}

	return user, nil
}

func checkInviteUser(c *ctx.ServiceContext, tx *db.Session, form *forms.InviteUserForm) (*models.User, e.Error) {
	var user *models.User
	var err e.Error

	if form.UserId != "" {
		user, err = getInviteUserById(c, tx, form.UserId)
		if err != nil {
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

	if user != nil && user.Id == consts.SysUserId {
		return nil, e.New(e.UserNotExists, fmt.Errorf("should not use sys user"), http.StatusBadRequest)
	}

	return user, nil

}

func createInviteUser(c *ctx.ServiceContext, tx *db.Session, form *forms.InviteUserForm, user *models.User, initPass string) (*models.User, bool, e.Error) {
	isNew := false

	hashedPassword, err := services.HashPassword(initPass)
	if err != nil {
		c.Logger().Errorf("error hash password, err %s", err)
		return user, isNew, err
	}
	if user == nil {
		user, err = services.CreateUser(tx, models.User{
			Name:     form.Name,
			Password: hashedPassword,
			Email:    form.Email,
			Phone:    form.Phone,
		})
		if err != nil && err.Code() == e.UserAlreadyExists {
			return user, isNew, e.New(err.Code(), err, http.StatusBadRequest)
		} else if err != nil {
			_ = tx.Rollback()
			c.Logger().Errorf("error create user, err %s", err)
			return user, isNew, err
		}
		isNew = true
	}

	return user, isNew, nil
}

func createInviteUserOrgRel(c *ctx.ServiceContext, tx *db.Session, form *forms.InviteUserForm, user *models.User, isNew bool) e.Error {
	if !isNew {
		if err := services.DeleteUserOrgRel(tx, user.Id, form.Id); err != nil {
			_ = tx.Rollback()
			c.Logger().Errorf("error del user org rel, err %s", err)
		}
	}
	if _, err := services.CreateUserOrgRel(tx, models.UserOrg{
		OrgId:  form.Id,
		UserId: user.Id,
		Role:   form.Role,
	}); err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("error create user org rel, err %s", err)
		return err
	}

	// 新用户自动加入演示组织
	if isNew && c.OrgId != models.Id(common.DemoOrgId) {
		if err := services.TryAddDemoRelation(tx, user.Id); err != nil {
			_ = tx.Rollback()
			c.Logger().Errorf("error add user demo rel, err %s", err)
			return err
		}
	}

	return nil
}

// InviteUser 邀请用户加入某个组织
// 如果用户不存在，则创建并加入组织，如果用户已经存在，则加入该组织
func InviteUser(c *ctx.ServiceContext, form *forms.InviteUserForm) (*resps.UserWithRoleResp, e.Error) {
	c.AddLogField("action", fmt.Sprintf("invite user %s%s to org %s as %s", form.Name, form.UserId, form.Id, form.Role))

	org, err := getInviteUserOrg(c, form)
	if err != nil {
		return nil, err
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
	user, err := checkInviteUser(c, tx, form)
	if err != nil {
		return nil, err
	}

	initPass := utils.GenPasswd(6, "mix")
	user, isNew, err := createInviteUser(c, tx, form, user, initPass)
	if err != nil {
		return nil, err
	}

	// 建立用户与组织间关联
	// 新用户自动加入演示组织
	err = createInviteUserOrgRel(c, tx, form, user, isNew)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("error commit invite user, err %s", err)
		return nil, e.New(e.DBError, err)
	}

	// 发送邀请邮件
	go sendInviteUserNotify(user, c.Username, org.Name, initPass, isNew)

	resp := resps.UserWithRoleResp{
		User: *user,
		Role: form.Role,
	}

	return &resp, nil
}

func SearchOrgResourcesFilters(c *ctx.ServiceContext, form *forms.SearchOrgResourceForm) (*resps.OrgEnvAndProviderResp, e.Error) {
	query := services.GetOrgResourcesQuery(c.DB().Model(&models.Resource{}), form.Q, c.OrgId, c.UserId, c.IsSuperAdmin)
	type SearchResult struct {
		EnvName  string    `json:"env_name"`
		EnvId    models.Id `json:"env_id"`
		Provider string    `json:"provider"`
	}
	rs := make([]SearchResult, 0)
	if err := query.Scan(&rs); err != nil {
		return nil, e.New(e.DBError, err)
	}
	r := &resps.OrgEnvAndProviderResp{}
	temp := map[string]interface{}{}
	for _, v := range rs {
		if _, ok := temp[v.EnvName]; !ok {
			// 通过map 对环境名称进行过滤
			r.Envs = append(r.Envs, resps.EnvResp{EnvName: v.EnvName, EnvId: v.EnvId})
			temp[v.EnvName] = nil
		}
		r.Providers = append(r.Providers, path.Base(v.Provider))
	}
	r.Providers = utils.Set(r.Providers)

	return r, nil
}

func SearchOrgResources(c *ctx.ServiceContext, form *forms.SearchOrgResourceForm) (interface{}, e.Error) {
	query := services.GetOrgResourcesQuery(c.DB().Model(&models.Resource{}), form.Q, c.OrgId, c.UserId, c.IsSuperAdmin)
	if len(form.EnvIds) != 0 {
		query = query.Where("iac_env.id in (?)", strings.Split(form.EnvIds, ","))
	}
	if len(form.Providers) != 0 {
		var tempSql []string
		var tempList []interface{}
		for _, v := range strings.Split(form.Providers, ",") {
			tempSql = append(tempSql, "iac_resource.provider like ?")
			tempList = append(tempList, strings.Join([]string{"%/", v}, ""))
		}
		query = query.Where(strings.Join(tempSql, " OR "), tempList...)
	}
	rs := make([]resps.OrgResourcesResp, 0)
	query = query.Order("project_id, env_id, provider desc")
	p := page.New(form.CurrentPage(), form.PageSize(), query)
	if err := p.Scan(&rs); err != nil {
		return nil, e.New(e.DBError, err)
	}
	for i := range rs {
		rs[i].Provider = path.Base(rs[i].Provider)
	}
	return &page.PageResp{
		Total:    p.MustTotal(),
		PageSize: p.Size,
		List:     rs,
	}, nil

}

// UpdateUserOrg 更新组织用户信息
func UpdateUserOrg(c *ctx.ServiceContext, form *forms.UpdateUserOrgForm) (userResp *resps.UserWithRoleResp, err e.Error) {
	c.AddLogField("action", fmt.Sprintf("update user %s in org %s to role %s", form.UserId, c.OrgId, form.Role))

	tx := c.Tx()

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()
	query := tx
	query = query.Where("status = 'enable'")
	if !c.IsSuperAdmin {
		userIds, _ := services.GetUserIdsByOrg(tx, c.OrgId)
		query = query.Where(fmt.Sprintf("%s.id in (?)", models.User{}.TableName()), userIds)
	}
	user, err := services.GetUserById(query, form.UserId)
	if err != nil && err.Code() == e.UserNotExists {
		return nil, e.New(err.Code(), err, http.StatusBadRequest)
	} else if err != nil {
		c.Logger().Errorf("error get user by id, err %s", err)
		return nil, e.New(e.DBError, err)
	}

	if err := services.UpdateUserOrgRel(tx, models.UserOrg{OrgId: c.OrgId, UserId: form.UserId, Role: form.Role}); err != nil {
		c.Logger().Errorf("error create user org rel, err %s", err)
		return nil, err
	}
	c.Logger().Infof("add user ", form.UserId, " to org ", c.OrgId, " succeed")

	attrs := models.Attrs{}
	if form.HasKey("name") {
		attrs["name"] = form.Name
	}
	if form.HasKey("phone") {
		attrs["phone"] = form.Phone
	}

	user, err = services.UpdateUser(tx, user.Id, attrs)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, e.New(e.DBError, err)
	}

	resp := resps.UserWithRoleResp{
		User: *user,
		Role: form.Role,
	}

	return &resp, nil
}

func sendInviteUserNotify(user *models.User, inviter, orgName, initPass string, isNew bool) {
	data := emailInviteUserData{
		User:         user,
		IsNewUser:    isNew,
		Inviter:      inviter,
		Organization: orgName,
		InitPass:     initPass,
		Addr:         configs.Get().Portal.Address,
	}
	err := mail.SendMail([]string{user.Email}, emailSubjectInviteUser, utils.SprintTemplate(consts.IacUserInvitationsTpl, data))
	if err != nil {
		logs.Get().Errorf("error send mail to %s, err %s", user.Email, err)
	}
}

// InviteUsersBatch 邀请多个用户加入某个组织
func InviteUsersBatch(c *ctx.ServiceContext, form *forms.InviteUsersBatchForm) (interface{}, e.Error) {
	var (
		success int
		failed  int
	)
	for _, v := range form.Email {
		f := forms.InviteUserForm{
			BaseForm: form.BaseForm,
			Id:       form.Id,
			Name:     v,
			Email:    v,
			Role:     form.Role,
		}
		if _, err := InviteUser(c, &f); err != nil {
			failed++
			c.Logger().Errorf("invite user err:%v", err)
		}
		success++
	}

	return resps.InviteUsersBatchResp{Success: success, Failed: failed}, nil
}

// OrgProjectsStat 组织和项目概览页统计数据
func OrgProjectsStat(c *ctx.ServiceContext, form *forms.OrgProjectsStatForm) (interface{}, e.Error) {
	if form.Limit == 0 {
		form.Limit = 10
	}
	tx := c.DB()
	// 环境状态占比
	envStat, err := services.GetOrgProjectsdEnvStat(tx, c.OrgId, form.ProjectIds)
	if err != nil {
		return nil, err
	}

	// 资源类型占比
	resStat, err := services.GetOrgProjectsResStat(tx, c.OrgId, form.ProjectIds, form.Limit)
	if err != nil {
		return nil, err
	}

	// 项目资源数量
	projectStat, err := services.GetOrgProjectStat(tx, c.OrgId, form.ProjectIds, form.Limit)
	if err != nil {
		return nil, err
	}

	// 资源新增趋势
	resGrowTrend, err := services.GetOrgResGrowTrend(tx, c.OrgId, form.ProjectIds)
	if err != nil {
		return nil, err
	}

	return &resps.OrgProjectsStatResp{
		EnvStat:      envStat,
		ResStat:      resStat,
		ProjectStat:  projectStat,
		ResGrowTrend: resGrowTrend,
	}, nil
}
