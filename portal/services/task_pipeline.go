// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package services

import (
	"cloudiac/common"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/utils/logs"
	"path/filepath"
)

func GetTplPipeline(sess *db.Session, tplId models.Id, revision, workdir string) (pipeline string, er e.Error) {
	repo, er := GetVcsRepoByTplId(sess, tplId)
	if er != nil {
		return pipeline, er
	}

	paths := []string{filepath.Join(workdir, common.PipelineFileName)}
	if workdir != "" {
		paths = append(paths, common.PipelineFileName)
	}

	var (
		content []byte
		err     error
	)
	for _, path := range paths {
		content, err = repo.ReadFileContent(revision, path)
		if err != nil {
			// TODO 所有 vcs 的 ReadFileContent() 实现需要在文件不存在时返回 ObjectNotExists 错误
			if e.Is(err, e.ObjectNotExists) {
				continue
			}

			logs.Get().Warnf("read file content error(%T): %v", err, err)
			return pipeline, e.New(e.VcsError, err)
		} else {
			break
		}
	}

	if len(content) == 0 {
		return "", nil
	}

	if _, err := DecodePipeline(string(content)); err != nil {
		return "", e.AutoNew(err, e.InvalidPipeline)
	}

	return string(content), nil
}

// 从 pipeline 中返回指定 typ 的 task，如果 pipeline 中未定义该类型 task 则返回默认 pipeline 中的值
func GetTaskFlowWithPipeline(p models.IPipeline, typ string) models.PipelineTaskFlow {
	defaultPipeline := models.MustGetPipelineByVersion(models.DefaultPipelineVersion)

	flow := defaultPipeline.GetTaskFlowWithPipeline(typ)
	customFlow := p.GetTaskFlowWithPipeline(typ)
	if customFlow.Image != "" {
		flow.Image = customFlow.Image
	}
	if len(customFlow.Steps) != 0 {
		flow.Steps = customFlow.Steps
	}
	if customFlow.OnFail != nil {
		flow.OnFail = customFlow.OnFail
	}
	if customFlow.OnSuccess != nil {
		flow.OnSuccess = customFlow.OnSuccess
	}
	return flow
}

func DecodePipeline(s string) (models.IPipeline, error) {
	defaultPipeline := models.DefaultPipeline()
	if s == "" {
		return defaultPipeline, nil
	}

	p := models.Pipeline{}
	ver, err := p.GetVersion(s)
	if err != nil {
		return nil, err
	}

	if ver == "0.3" || ver == "0.4" {
		return models.NewPipelineDot34(s)
	}

	if ver == "0.5" {
		return models.NewPipelineDot5(s)
	}

	return nil, e.New(e.InvalidPipelineVersion)
}

func UpdateTaskContainerId(sess *db.Session, taskId models.Id, containerId string) e.Error {
	task := &models.Task{}
	task.ContainerId = containerId
	_, err := models.UpdateModel(sess, task, "id = ?", taskId)
	if err != nil {
		return e.AutoNew(err, e.DBError)
	}
	return nil
}

func UpdateScanTaskContainerId(sess *db.Session, taskId models.Id, containerId string) e.Error {
	task := &models.ScanTask{}
	task.ContainerId = containerId
	_, err := models.UpdateModel(sess, task, "id = ?", taskId)
	if err != nil {
		return e.AutoNew(err, e.DBError)
	}
	return nil
}
