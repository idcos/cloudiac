// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.
package main

import (
	"cloudiac/configs"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
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

	envs, err := searchEnvs(tx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	runners, err := services.RunnerSearch()
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	runnerIds := make(map[string]string)
	for _, runner := range runners {
		runnerIds[runner.ID] = strings.Join(runner.Tags, ",")
	}

	if err := updateRunnerIdToRunnerTags(tx, envs, runnerIds); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return err
	}
	return nil
}

func searchEnvs(query *db.Session) ([]*models.Env, error) {
	envs := make([]*models.Env, 0)
	query = services.QueryEnv(query)
	if err := query.Find(&envs); err != nil {
		return nil, fmt.Errorf("database error")
	}
	return envs, nil
}

func updateRunnerIdToRunnerTags(tx *db.Session, envs []*models.Env, runnerIds map[string]string) error {
	attrs := models.Attrs{}
	var noArchivedEnvs []*models.Env
	for _, env := range envs {
		if !env.Archived && env.RunnerId != "" {
			noArchivedEnvs = append(noArchivedEnvs, env)
		}
	}
	if len(noArchivedEnvs) == 0 {
		return fmt.Errorf("no env runner-id needs to be modified")
	}
	for _, noArchivedEnv := range noArchivedEnvs {
		for runnerId, runnerTag := range runnerIds {
			if noArchivedEnv.RunnerId != runnerId {
				continue
			}
			attrs["runner_tags"] = runnerTag
			attrs["runner_id"] = ""
			if _, err := tx.Model(&models.Env{}).Where("id = ?", noArchivedEnv.Id).UpdateAttrs(attrs); err != nil {
				panic("update has been fail")
			}
			logger.Infof("change runnerId to runnerTags success, envId = %s,runnerId=%s,runnerTags = %s", noArchivedEnv.Id, runnerId, runnerTag)
		}
	}
	return nil
}
