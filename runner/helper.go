// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package runner

import (
	"cloudiac/common"
	"cloudiac/configs"
	"cloudiac/utils"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/errdefs"
	"github.com/pkg/errors"
)

type IaCTemplate struct {
	TemplateUUID string
	TaskId       string
}

// ReqBody from reqeust
type ReqBody struct {
	Repo         string     `json:"repo"`
	RepoCommit   string     `json:"repo_commit"`
	RepoRevision string     `json:"repo_revision"`
	TemplateUUID string     `json:"template_uuid"`
	TaskID       string     `json:"task_id"`
	DockerImage  string     `json:"docker_image" defalut:"mt5225/tf-ansible:v0.0.1"`
	StateStore   StateStore `json:"state_store"`
	Env          map[string]string
	Timeout      int    `json:"timeout" default:"600"`
	Mode         string `json:"mode" default:"plan"`
	Varfile      string `json:"varfile"`
	Extra        string `json:"extra"`
	Playbook     string `json:"playbook" form:"playbook" `

	PrivateKey string `json:"privateKey"`
}

type SchemaValueDetail struct {
	Sensitive bool `json:"sensitive,omitempty"`
}

type BlockDetail struct {
	Attributes map[string]SchemaValueDetail `json:"attributes"`
}

type ResourceSchemas struct {
	Block BlockDetail `json:"block"`
}

type Schemas struct {
	ResourceSchemas map[string]ResourceSchemas `json:"resource_schemas"`
}

type ProviderMeta struct {
	ProviderSchemas map[string]Schemas `json:"provider_schemas"`
}

type ProviderSensitiveAttrMap map[string][]string

// PathExists 判断目录是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// GetEnvPluginCache 获取环境私有的plugin-cache目录
func GetEnvPluginCache(envId string) string {
	conf := configs.Get()
	return filepath.Join(conf.Runner.AbsStoragePath(), envId)
}

func GetTaskWorkspace(envId string, taskId string) string {
	conf := configs.Get()
	return filepath.Join(conf.Runner.AbsStoragePath(), envId, taskId)
}

func GetTaskDir(envId string, taskId string, step int) string {
	return filepath.Join(GetTaskWorkspace(envId, taskId), GetTaskDirName(step))
}

func GetTaskDirName(step int) string {
	if step == common.CollectTaskStepIndex {
		return ".step-collect"
	}
	return fmt.Sprintf("step%d", step)
}

func FetchTaskLog(envId string, taskId string, step int) ([]byte, error) {
	path := filepath.Join(GetTaskDir(envId, taskId, step), TaskLogName)
	return ioutil.ReadFile(path)
}

func FetchStateJson(envId string, taskId string) ([]byte, error) {
	path := filepath.Join(GetTaskWorkspace(envId, taskId), TFStateJsonFile)
	content, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return content, nil
}

func FetchProviderJson(envId string, taskId string) ([]byte, error) {
	path := filepath.Join(GetTaskWorkspace(envId, taskId), TFProviderSchema)
	content, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	providerSchemaContent, err := BuildProviderSensitiveAttrMap(content)
	if err != nil {
		return nil, err
	}
	return providerSchemaContent, nil
}

func BuildProviderSensitiveAttrMap(body []byte) ([]byte, error) {
	providerMeta := &ProviderMeta{}
	err := json.Unmarshal(body, providerMeta)
	if err != nil {
		return nil, err
	}
	proMap := ProviderSensitiveAttrMap{}
	for provider, v := range providerMeta.ProviderSchemas {
		for resourceType, value := range v.ResourceSchemas {
			keys := []string{}
			for attrKey, attrValue := range value.Block.Attributes {
				if attrValue.Sensitive {
					keys = append(keys, attrKey)
				}
			}
			if len(keys) != 0 {
				proMap[strings.Join([]string{provider, resourceType}, "-")] = keys
			}
		}

	}

	providerMap, err := json.Marshal(proMap)
	if err != nil {
		return nil, err
	}
	return providerMap, nil
}

func FetchPlanJson(envId string, taskId string) ([]byte, error) {
	path := filepath.Join(GetTaskWorkspace(envId, taskId), TFPlanJsonFile)
	content, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return content, nil
}

func FetchJson(envId string, taskId string, jsonFile string) ([]byte, error) {
	path := filepath.Join(GetTaskWorkspace(envId, taskId), jsonFile)
	content, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return content, nil
}

func GetLatestStepInfo(envId string, taskId string) (info StepInfo, err error) {
	path := filepath.Join(GetTaskWorkspace(envId, taskId), TaskStepInfoFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return info, nil
		}
		return info, err
	}
	if err := json.Unmarshal(data, &info); err != nil {
		return info, err
	}
	return info, nil
}

func WriteTaskControlInfo(info TaskControlInfo) error {
	path := TaskControlFilePath(info.EnvId, info.TaskId)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	logger.Debugf(">>> @feat-task-abort write task control info to: %v, info=%#v", path, info)
	return os.WriteFile(path, utils.MustJSON(info), 0600)
}

func ReadTaskControlInfo(envId, taskId string) (info TaskControlInfo, err error) {
	data, err := os.ReadFile(TaskControlFilePath(envId, taskId))
	if err != nil {
		if os.IsNotExist(err) {
			return info, nil
		}
		return info, err
	}

	if err := json.Unmarshal(data, &info); err != nil {
		return info, err
	}
	return info, nil
}

func TaskControlFilePath(envId, taskId string) string {
	return filepath.Join(GetTaskWorkspace(envId, taskId), TaskControlFileName)
}

func KillContainers(ctx context.Context, cids ...string) error {
	cli, err := DockerClient()
	if err != nil {
		return err
	}

	// 这里仅 kill container，container 的删除通过启动时的 AutoRemove 参数配置
	for _, cid := range cids {
		// default signal "SIGKILL"
		if err := cli.ContainerKill(ctx, cid, ""); err != nil {
			var targetErr errdefs.ErrNotFound
			if errors.As(err, &targetErr) {
				continue
			}

			// 有可能己经提交了删除请求，这里忽略掉这些报错
			if !strings.Contains(err.Error(), "already in progress") &&
				!strings.Contains(err.Error(), "No such container") {
				logger.Info("kill container error: %v", err)
				continue
			}
			return err
		}
	}
	return nil
}

//判断provider缓存目录是否存在，存在删除
func DeleteProviderCache(host, source, version string) (ok bool, err error) {
	fullPath := filepath.Join(host, source, version)
	exist, err := PathExists(fullPath)
	if err == nil && exist {
		err := os.RemoveAll(fullPath)
		if err != nil {
			return false, err
		}
	} else {
		return false, err
	}
	return true, nil
}
