package services

import (
	"cloudiac/configs"
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/utils"
	"cloudiac/utils/logs"
	"cloudiac/utils/mail"
	"fmt"
	"strings"
)

func CreateUserDemoOrgData(c *ctx.ServiceContext, tx *db.Session, user *models.User) e.Error {
	logger := c.Logger().WithField("user", user.Email)
	userName := user.Name
	if userName == "" {
		userName = strings.Split(user.Email, "@")[0]
	}

	var (
		org     *models.Organization
		vcs     *models.Vcs
		project *models.Project
		er      e.Error
	)

	// 尝试获取一个可用的组织名称
	for i := 0; i < 80; i++ { // 限制重试次数
		demoName := fmt.Sprintf("%s的演示组织", userName)
		if i > 0 { // 第一次尝试不加随机后缀，此后都加至少两位的随机后缀(每尝试10次后缀长度加1)
			demoName += utils.RandomStr(i/10 + 2)
		}
		logger.Debugf("i=%v, demo name: %v", i, demoName)
		org, er = CreateOrganization(tx, models.Organization{
			Name:        demoName,
			Description: "",
			CreatorId:   user.Id,
			IsDemo:      true,
		})
		if er == nil {
			break
		} else if er.Code() == e.OrganizationAlreadyExists {
			continue
		} else {
			return er
		}
	}
	logger.Infof("create demo org: %s(%s)", org.Name, org.Id)

	// 创建 demo vcs
	demoCfg := configs.Get().Demo
	vcs, er = CreateDemoVcs(tx, org.Id, demoCfg.VcsType, demoCfg.VcsAddress, demoCfg.VcsName, demoCfg.VcsToken)
	if er != nil {
		return er
	}
	logger.Infof("create demo vcs: %s(%s)", vcs.Name, vcs.Id)

	// 创建 demo project
	project, er = CreateProject(tx, &models.Project{
		Name:        demoCfg.ProjectName,
		OrgId:       org.Id,
		Description: "",
		CreatorId:   user.Id,
		IsDemo:      true,
	})
	if er != nil {
		return er
	}
	logger.Infof("create demo project: %s(%s)", project.Name, project.Id)

	// // 创建 demo tpl
	for _, t := range demoCfg.Templates {
		tpl, er := CreateDemoTemplate(tx, org.Id, vcs.Id, project.Id, user.Id, t)
		if er != nil {
			return er
		}
		logger.Infof("create demo template: %s(%s)", tpl.Name, tpl.Id)
	}

	logger.Infof("grant %s role to user", consts.OrgRoleAdmin)
	// 当前用户自动添加为组织管理员
	if _, er = CreateUserOrgRel(tx, models.UserOrg{
		OrgId:  org.Id,
		UserId: user.Id,
		Role:   consts.OrgRoleAdmin,
	}); er != nil {
		return er
	}

	return nil
}

func CreateDemoTemplate(tx *db.Session, orgId, vcsId, projectId, userId models.Id, t configs.DemoTemplate) (*models.Template, e.Error) {
	tpl, er := CreateTemplate(tx, models.Template{
		Name:         t.Name,
		OrgId:        orgId,
		Description:  "",
		VcsId:        vcsId,
		RepoId:       t.RepoId,
		RepoFullName: "",
		RepoRevision: t.Revison,
		CreatorId:    userId,
		Workdir:      "",
		TfVarsFile:   t.TfVars,
		IsDemo:       true,
	})
	if er != nil {
		return nil, er
	}

	// 创建模板与项目的关系
	if err := CreateTemplateProject(tx, []models.Id{projectId}, tpl.Id); err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	vars := make([]models.Variable, 0)
	for _, tv := range t.Variables {
		mv := models.Variable{
			OrgId: orgId,
			TplId: tpl.Id,
			VariableBody: models.VariableBody{
				Name:        tv.Name,
				Value:       tv.Value,
				Scope:       consts.ScopeTemplate,
				Type:        consts.VarTypeEnv,
				Description: tv.Description,
				Sensitive:   tv.Sensitive,
			},
		}
		mv.Id = models.NewId("var")
		vars = append(vars, mv)
	}

	_, er = UpdateObjectVars(tx, consts.ScopeTemplate, tpl.Id, vars)
	if er != nil {
		return nil, er
	}
	return tpl, nil
}

func CreateDemoVcs(tx *db.Session, orgId models.Id, typ, addr, name, token string) (*models.Vcs, e.Error) {
	token, err := utils.EncryptSecretVar(token)
	if err != nil {
		return nil, e.New(e.VcsError, err)
	}
	v := models.Vcs{
		OrgId:    orgId,
		VcsType:  typ,
		Name:     name,
		Address:  addr,
		VcsToken: token,
		IsDemo:   true,
	}

	vcs, err := CreateVcs(tx, v)
	if err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}
	return vcs, nil
}

func SendActivateAccountMail(user *models.User, token string) e.Error {
	name := user.Name
	if name == "" {
		name = strings.Split(user.Email, "@")[0]
	}
	content := utils.SprintTemplate(consts.UserActivteMail, map[string]interface{}{
		"Name":    name,
		"Email":   user.Email,
		"Address": utils.JoinURL(configs.Get().Portal.Address, "register/activation/", token),
	})

	er := mail.SendMail([]string{user.Email}, "欢迎注册 CloudIac", content)
	if er != nil {
		logs.Get().Errorf("error send mail to %s, err %s", user.Email, er)
		return er
	}
	return nil
}
