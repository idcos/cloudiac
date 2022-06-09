// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package services

import (
	"cloudiac/configs"
	"cloudiac/portal/models"
	"cloudiac/portal/services/logstorage"
	"cloudiac/utils"
	"cloudiac/utils/logs"
	"fmt"
	"net/http"
	"path/filepath"
	"time"
)

type syncResRequest struct {
	Source    string `json:"source"`
	OrgId     string `json:"org_id"`
	ProjectId string `json:"project_id"`
	EnvId     string `json:"env_id"`
	TaskId    string `json:"task_id"`

	Resources []syncResource `json:"resources"`
}

type syncResource struct {
	ResourceType string `json:"resource_type"`
	ResourceId   string `json:"resource_id"`
	OperateType  string `json:"operate_type"`
	OperateTime  string `json:"operate_time"`
}

func syncManagedResToProvider(task *models.Task) {
	if configs.Get().AlicloudResSyncApi == "" {
		return
	}

	logger := logs.Get().WithField("taskId", task.Id).WithField("envId", task.EnvId)
	var (
		tfState   *TfState
		tfPlan    *TfPlan
		envResMap = make(map[string]*models.Resource)
	)

	{
		content, err := logstorage.Get().Read(task.PlanJsonPath())
		if err != nil {
			logger.Errorf("read %s: %v", task.PlanJsonPath(), err)
			return
		}
		tfPlan, err = UnmarshalPlanJson(content)
		if err != nil {
			logger.Warnf("unmarshal plan json: %v", err)
			return
		}
	}

	{
		content, err := logstorage.Get().Read(task.StateJsonPath())
		if err != nil {
			logger.Errorf("read %s: %v", task.StateJsonPath(), err)
			return
		}

		tfState, err = UnmarshalStateJson(content)
		if err != nil {
			logger.Warnf("unmarshal state json: %v", err)
			return
		}

		rs := make([]*models.Resource, 0)
		rs = append(rs, traverseStateModule(&tfState.Values.RootModule)...)
		for i := range tfState.Values.ChildModules {
			rs = append(rs, traverseStateModule(&tfState.Values.ChildModules[i])...)
		}
		for _, r := range rs {
			envResMap[r.Address] = r
		}
	}

	syncRess := make([]syncResource, 0)
	operateTime := time.Time(task.UpdatedAt).Format(time.RFC3339)
	for _, r := range tfPlan.ResourceChanges {
		providerName := filepath.Base(r.ProviderName)
		if providerName != "alicloud" {
			continue
		}

		actions := r.Change.Actions
		if utils.StrInArray("create", actions...) {
			envRes, ok := envResMap[r.Address]
			if !ok {
				logger.Warnf("environment resource not found, address %s", r.Address)
				continue
			}
			attrs := envRes.Attrs
			sr := syncResource{
				ResourceType: r.Type,
				OperateTime:  operateTime,
			}
			sr.ResourceId = firstMapStrVal(attrs, "id")
			sr.OperateType = "create"
			syncRess = append(syncRess, sr)
		}

		if utils.StrInArray("delete", actions...) {
			before, ok := r.Change.Before.(map[string]interface{})
			if !ok {
				logger.Warnf("invalid resource_changes[].change.before: %v", r.Change.Before)
				continue
			}

			sr := syncResource{
				ResourceType: r.Type,
				OperateTime:  operateTime,
			}
			sr.ResourceId = firstMapStrVal(before, "id")
			sr.OperateType = "delete"
			syncRess = append(syncRess, sr)
		}
	}

	err := syncManagedResToAlicloud(syncResRequest{
		Source:    "cloudiac",
		OrgId:     task.OrgId.String(),
		ProjectId: task.ProjectId.String(),
		EnvId:     task.EnvId.String(),
		TaskId:    task.Id.String(),
		Resources: syncRess,
	})
	if err != nil {
		logger.Errorf("%v", err)
	}
}

func syncManagedResToAlicloud(req syncResRequest) error {
	if len(req.Resources) <= 0 {
		return nil
	}

	header := make(http.Header)
	header.Add("content-type", "application/json")
	apiUrl := configs.Get().AlicloudResSyncApi
	resp, err := utils.HttpService(apiUrl, "POST", &header, req, 10, 30)
	if err != nil {
		return err
	}
	logs.Get().Infof("sync managed resources response: %s", resp)
	return nil
}

func firstMapStrVal(m map[string]interface{}, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			s := fmt.Sprintf("%v", v)
			if s != "" {
				return s
			}
		}
	}
	return ""
}
