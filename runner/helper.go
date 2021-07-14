package runner

import (
	"cloudiac/configs"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
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
	return fmt.Sprintf("step%d", step)
}

func FetchTaskStepLog(envId string, taskId string, step int) ([]byte, error) {
	path := filepath.Join(GetTaskStepDir(envId, taskId, step), TaskLogName)
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
