// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.
package main

import (
	"cloudiac/configs"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/portal/services"
	"cloudiac/utils"
	"fmt"
	"strings"
)

// ./iac-tool upgrade2v0.10

type Update2v0dot10Cmd struct{}

func (*Update2v0dot10Cmd) Execute(args []string) error {
	logger.Infof("upgrade to v0.10 ...")

	configs.Init(opt.Config)
	db.Init(configs.Get().Mysql)
	models.Init(true)

	return db.Get().Transaction(func(tx *db.Session) error {
		envs, err := searchEnvs(tx)
		if err != nil {
			return err
		}

		runners, err := services.RunnerSearch()
		if err != nil {
			return err
		}

		logger.Infof("find runners: %s", utils.MustJSON(runners))
		if len(runners) <= 0 {
			return fmt.Errorf("no runner services, please start ct-runner first")
		}

		runnerIds := make(map[string]string)
		for _, runner := range runners {
			runnerIds[runner.ID] = strings.Join(runner.Tags, ",")
		}

		if err := updateRunnerIdToRunnerTags(tx, envs, runnerIds); err != nil {
			return err
		}

		sqls := []string{
			`update iac_resource SET res_id=JSON_UNQUOTE(JSON_EXTRACT(attrs, "$.id")) where res_id = ''`,
			`update iac_resource join (select iac_task.id, iac_task.end_at from iac_task join (
				select env_id,max(end_at) as end_at from iac_task group by env_id) t 
				on t.env_id = iac_task.env_id and t.end_at = iac_task.end_at
			) tt on tt.id = iac_resource.task_id set iac_resource.applied_at=tt.end_at`,
			`UPDATE iac_resource, ( select env_id,res_id,min(applied_at) as applied_at from iac_resource 
				where applied_at is not NULL group by env_id,res_id) t 
			SET iac_resource.applied_at=t.applied_at 
			WHERE t.env_id = iac_resource.env_id AND t.res_id = iac_resource.res_id;`,
		}
		for _, sql := range sqls {
			if _, err := tx.Debug().Exec(sql); err != nil {
				return err
			}
		}

		return nil
	})
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

	replaces := make([]*models.Env, 0)
	missed := make([]*models.Env, 0)
	for _, noArchivedEnv := range noArchivedEnvs {
		if runnerTag, ok := runnerIds[noArchivedEnv.RunnerId]; ok {
			attrs["runner_tags"] = runnerTag
			attrs["runner_id"] = ""
			if _, err := tx.Model(&models.Env{}).Where("id = ?", noArchivedEnv.Id).UpdateAttrs(attrs); err != nil {
				panic("update has been fail")
			}
			logger.Infof("replace runner id to tags success, envId=%s, runnerId=%s, runnerTags=%s",
				noArchivedEnv.Id, noArchivedEnv.RunnerId, runnerTag)
			replaces = append(replaces, noArchivedEnv)
		} else {
			missed = append(missed, noArchivedEnv)
		}
	}

	for _, e := range missed {
		logger.Warnf("runner not found, envId=%s, runnerId=%s", e.Id, e.RunnerId)
	}
	logger.Infof("replace runner id to tags summary, replaced: %d, missed: %d", len(replaces), len(missed))
	return nil
}
