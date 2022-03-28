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

	tx := db.Get().Begin()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	envDetails, err := searchEnvDetails(tx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	runners, err := services.RunnerSearch()
	if err != nil {
		return err
	}

	runnerIds := make(map[string]string)
	for _, runner := range runners {
		runnerIds[runner.ID] = strings.Join(runner.Tags, ",")
	}
	if err := updateRunnerIdToRunnerTags(tx, envDetails, runnerIds); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return err
	}
	return nil
}

func searchEnvDetails(query *db.Session) ([]*models.EnvDetail, error) {
	query = services.QueryEnvDetail(query)
	form := forms.PageForm{}
	p := page.New(form.CurrentPage(), form.PageSize(), query)

	details := make([]*models.EnvDetail, 0)
	if err := p.Scan(&details); err != nil {
		return nil, fmt.Errorf("database error")
	}
	return details, nil
}

func updateRunnerIdToRunnerTags(tx *db.Session, envDetails []*models.EnvDetail, runnerIds map[string]string) error {
	attrs := models.Attrs{}
	var noArchivedEnvs []*models.EnvDetail
	for _, env := range envDetails {
		if !env.Env.Archived && env.Env.RunnerId != "" {
			noArchivedEnvs = append(noArchivedEnvs, env)
		}
	}
	if len(noArchivedEnvs) == 0 {
		return fmt.Errorf("no env runner-id needs to be modified")
	}
	for _, noArchivedEnv := range noArchivedEnvs {
		for runnerId, runnerTag := range runnerIds {
			if noArchivedEnv.Env.RunnerId != runnerId {
				continue
			}
			attrs["runner_tags"] = runnerTag
			attrs["runner_id"] = ""
			if _, err := tx.Model(&models.Env{}).Where("id = ?", noArchivedEnv.Env.Id).UpdateAttrs(attrs); err != nil {
				panic("update has been fail")
			}
			logger.Infof("change runnerId to runnerTags success, envId = %s,runnerId=%s,runnerTags = %s", noArchivedEnv.Env.Id, runnerId, runnerTag)
		}
	}
	return nil
}
