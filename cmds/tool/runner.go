// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.
package main

import (
	"cloudiac/configs"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/libs/page"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
	"fmt"
	"strings"
)

// ./iac-tool update-env-id2tag

type ChangeRunnerIdToRunnerTagsCmd struct {
}

func (*ChangeRunnerIdToRunnerTagsCmd) Execute(args []string) error {
	configs.Init(opt.Config)
	db.Init(configs.Get().Mysql)
	models.Init(false)

	query := db.Get().Begin()
	tx := db.Get()
	query = services.QueryEnvDetail(query)

	form := forms.PageForm{}
	p := page.New(form.CurrentPage(), form.PageSize(), query)

	details := make([]*models.EnvDetail, 0)
	if err := p.Scan(&details); err != nil {
		return fmt.Errorf("database error")
	}
	runners, err := services.RunnerSearch()
	if err != nil {
		return err
	}

	tags := make([]string, len(runners))
	runnerIds := make([]string, len(runners))

	for _, runner := range runners {
		runnerIds = append(runnerIds, runner.ID)
		tags = append(tags, strings.Join(runner.Tags, ","))
	}

	attrs := models.Attrs{}
	var noArchivedEnvs []*models.EnvDetail
	for _, env := range details {
		if !env.Env.Archived && env.Env.RunnerId != "" {
			noArchivedEnvs = append(noArchivedEnvs, env)
		}
	}
	if len(noArchivedEnvs) == 0 {
		return fmt.Errorf("no env runner-id needs to be modified")
	}
	for _, noArchivedEnv := range noArchivedEnvs {
		for runnerIdIndex, runnerId := range runnerIds {
			if noArchivedEnv.Env.RunnerId != runnerId {
				continue
			}
			attrs["runner_tags"] = tags[runnerIdIndex]
			attrs["runner_id"] = ""
			if _, err := tx.Model(&models.Env{}).Where("id = ?", noArchivedEnv.Env.Id).UpdateAttrs(attrs); err != nil {
				return fmt.Errorf("update has been fail,envId = %s", noArchivedEnv.Env.Id)
			}
			logger.Infof("change runnerId to runnerTags success, envId = %s,runnerId=%s,runnerTags = %s", noArchivedEnv.Env.Id, runnerId, tags[runnerIdIndex])
		}
	}
	defer func() {
		if r := recover(); r != nil {
			_ = query.Rollback()
			_ = tx.Rollback()
			panic(r)
		}
	}()
	return nil
}
