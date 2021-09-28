// Copyright 2021 CloudJ Company Limited. All rights reserved.

package runner

import (
	"cloudiac/common"
	"cloudiac/configs"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
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

func GetTaskWorkspace(envId string, taskId string) string {
	conf := configs.Get()
	return filepath.Join(conf.Runner.AbsStoragePath(), envId, taskId)
}

func GetTaskStepDir(envId string, taskId string, step int) string {
	return filepath.Join(GetTaskWorkspace(envId, taskId), GetTaskStepDirName(step))
}

func GetTaskStepDirName(step int) string {
	if step == common.CollectTaskStepIndex {
		return fmt.Sprintf(".step-collect")
	}
	return fmt.Sprintf("step%d", step)
}

func FetchTaskStepLog(envId string, taskId string, step int) ([]byte, error) {
	path := filepath.Join(GetTaskStepDir(envId, taskId, step), TaskStepLogName)
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
