// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package main

import (
	"cloudiac/configs"
	"cloudiac/portal/consts"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
	"cloudiac/utils"
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

type Organization struct {
	Name        string     `yaml:"name"`
	Description string     `yaml:"description"`
	Project     Project    `yaml:"project"`
	VCS         VCS        `yaml:"vcs"`
	Key         Key        `yaml:"key"`
	Variables   []Variable `yaml:"variables"`
	Template    Template   `yaml:"template"`
}

type Project struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

type VCS struct {
	Name    string `yaml:"name"`
	Type    string `yaml:"type"`
	Address string `yaml:"address"`
	Token   string `yaml:"token"`
}

type Key struct {
	Name    string `yaml:"name"`
	KeyFile string `yaml:"key_file"`
}

type Variable struct {
	Name        string `yaml:"name"`
	Type        string `yaml:"type"`
	Value       string `yaml:"value"`
	Description string `yaml:"description"`
	Sensitive   bool   `yaml:"sensitive"`
}

type Template struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	VCS         string   `yaml:"vcs"`
	RepoId      string   `yaml:"repo_id"`
	Revision    string   `yaml:"revision"`
	TfVarsFile  string   `yaml:"tf_vars_file"`
	Playbook    string   `yaml:"playbook"`
	Variables   []string `yaml:"variables"`
	Workdir     string   `yaml:"workdir"`
}

type InitDemo struct {
	Organization Organization `yaml:"organization"`
}

var (
	data InitDemo
)

func (*InitDemo) Usage() string {
	return `<demo config file>`
}

func (p *InitDemo) Execute(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("demo config file is required")
	}

	configs.Init(opt.Config)
	db.Init(configs.Get().Mysql)
	models.Init(false)

	config := args[0]
	if err := parseConfig(config, &data); err != nil {
		panic("parse demo config failed")
	}
	fmt.Printf("Load from demo config: %s\n", config)

	// 创建演示组织
	tx := db.Get().Begin()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	o := models.Organization{
		Name:        data.Organization.Name,
		Description: data.Organization.Description,
		IsDemo:      true,
		CreatorId:   consts.SysUserId,
	}
	o.Id = consts.DemoOrgId

	org, err := services.CreateOrganization(tx, o)
	if err != nil {
		panic(fmt.Errorf("create org failed, err %s", err))
	}
	fmt.Printf("org_id = %s\n", org.Id)

	// 创建项目
	project, err := services.CreateProject(tx, &models.Project{
		OrgId:       org.Id,
		Name:        data.Organization.Project.Name,
		Description: data.Organization.Project.Description,
		CreatorId:   consts.SysUserId,
	})
	if err != nil {
		panic(fmt.Errorf("create project failed, err %s", err))
	}
	fmt.Printf("project_id = %s\n", project.Id)

	// 获取默认VCS
	vcs, er := services.GetDefaultVcs(tx)
	if er != nil {
		panic(fmt.Errorf("missing default vcs, err %s", er))
	}
	fmt.Printf("default_vcs_id = %s\n", vcs.Id)

	// 创建密钥
	content, er := ioutil.ReadFile(data.Organization.Key.KeyFile)
	if er != nil {
		panic(fmt.Errorf("read key file %s failed, err %s", data.Organization.Key.KeyFile, er))
	}
	encrypted, er := utils.AesEncrypt(string(content))
	if er != nil {
		panic(fmt.Errorf("encrypt key failed, err %s", er))
	}
	key, err := services.CreateKey(tx, models.Key{
		OrgId:     org.Id,
		Name:      data.Organization.Key.Name,
		Content:   encrypted,
		CreatorId: consts.SysUserId,
	})
	if err != nil {
		panic(fmt.Errorf("create key failed, err %s", err))
	}
	fmt.Printf("key_id = %s\n", key.Id)

	// 创建变量
	variables := make([]forms.Variable, 0)
	for _, v := range data.Organization.Variables {
		variables = append(variables, forms.Variable{
			Scope:       consts.ScopeOrg,
			Name:        v.Name,
			Description: v.Description,
			Value:       v.Value,
			Type:        v.Type,
			Sensitive:   v.Sensitive,
		})
	}
	if err := services.OperationVariables(tx, org.Id, "", "", "", variables, nil); err != nil {
		panic(fmt.Errorf("create variable failed, err %s", err))
	}

	// 创建模版
	template, err := services.CreateTemplate(tx, models.Template{
		OrgId:        org.Id,
		Name:         data.Organization.Template.Name,
		Description:  data.Organization.Template.Description,
		VcsId:        vcs.Id,
		RepoId:       data.Organization.Template.RepoId,
		RepoRevision: data.Organization.Template.Revision,
		TfVarsFile:   data.Organization.Template.TfVarsFile,
		Playbook:     data.Organization.Template.Playbook,
		Workdir:      data.Organization.Template.Workdir,
		CreatorId:    consts.SysUserId,
	})
	if err != nil {
		panic(fmt.Errorf("create template failed, err %s", err))
	}
	fmt.Printf("template_id = %s\n", template.Id)

	// 创建模板与项目的关系
	err = services.CreateTemplateProject(tx, []models.Id{project.Id}, template.Id)
	if err != nil {
		panic(fmt.Errorf("create template relation failed, err %s", err))
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		panic(fmt.Errorf("error commit database, err %s", err))
	}

	fmt.Println("Demo data init completed!")
	return nil
}

func parseConfig(filename string, out interface{}) error {
	bs, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	bs = []byte(os.ExpandEnv(string(bs)))

	if err := yaml.Unmarshal(bs, out); err != nil {
		return fmt.Errorf("yaml.Unmarshal: %v", err)
	}

	return nil
}
