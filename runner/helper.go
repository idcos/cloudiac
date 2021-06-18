package runner

import (
	"cloudiac/configs"
	"cloudiac/utils/logs"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

type IaCTemplate struct {
	TemplateUUID string
	TaskId       string
}

// ReqBody from reqeust
type ReqBody struct {
	Repo         string     `json:"repo"`
	RepoCommit   string     `json:"repo_commit"`
	RepoBranch   string     `json:"repo_branch"`
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

type CommitedTask struct {
	TemplateId  string `json:"templateId"`
	TaskId      string `json:"taskId"`
	ContainerId string `json:"containerId"`

	containerInfoLock sync.RWMutex `json:"-"`
}

// 判断目录是否存在
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

// 创建目录
func PathCreate(path string) error {
	pathExists, err := PathExists(path)
	if err != nil {
		return err
	}
	if pathExists == true {
		return nil
	} else {
		err := os.MkdirAll(path, os.ModePerm)
		return err
	}
}

func GetTaskWorkDir(templateUUID string, taskId string) string {
	conf := configs.Get()
	return filepath.Join(conf.Runner.AbsStoragePath(), templateUUID, taskId)
}

func FetchTaskLog(templateUUID string, taskId string) ([]byte, error) {
	logFile := filepath.Join(GetTaskWorkDir(templateUUID, taskId), TaskLogName)
	return ioutil.ReadFile(logFile)
}

func FetchStateList(templateUUID string, taskId string) ([]byte, error) {
	logFile := filepath.Join(GetTaskWorkDir(templateUUID, taskId), TerraformStateListName)
	return ioutil.ReadFile(logFile)
}

func MakeTaskWorkDir(tplId string, taskId string) (string, error) {
	workDir := GetTaskWorkDir(tplId, taskId)
	err := PathCreate(workDir)
	return workDir, err
}

func ReqToTask(req *http.Request) (*CommitedTask, error) {
	bodyData, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	logger := logs.Get()
	logger.Debugf("task status request: %s", bodyData)

	var d CommitedTask
	if err := json.Unmarshal(bodyData, &d); err != nil {
		return nil, err
	}
	return &d, nil
}

// ReqToCommand create command structure to run container
// from POST request
func ReqToCommand(req *http.Request) (*Command, *StateStore, error) {
	bodyData, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, nil, err
	}

	logger := logs.Get()
	logger.Tracef("new task request: %s", bodyData)
	var d ReqBody
	if err := json.Unmarshal(bodyData, &d); err != nil {
		return nil, nil, err
	}

	c := new(Command)

	state := d.StateStore

	if d.DockerImage == "" {
		conf := configs.Get()
		c.Image = conf.Runner.DefaultImage
	} else {
		c.Image = d.DockerImage
	}

	for k, v := range d.Env {
		c.Env = append(c.Env, fmt.Sprintf("%s=%s", k, v))
	}

	for k, v := range AnsibleEnv {
		c.Env = append(c.Env, fmt.Sprintf("%s=%s", k, v))
	}

	workingDir, err := MakeTaskWorkDir(d.TemplateUUID, d.TaskID)
	if err != nil {
		return nil, nil, err
	}

	c.PrivateKey = d.PrivateKey

	c.TaskWorkdir = workingDir
	scriptPath := filepath.Join(c.TaskWorkdir, TaskScriptName)
	if err := GenScriptContent(&d, scriptPath); err != nil {
		return nil, nil, err
	}

	containerScriptPath := filepath.Join(ContainerTaskDir, TaskScriptName)
	containerLogPath := filepath.Join(ContainerTaskDir, TaskLogName)
	c.Commands = []string{"sh", "-c", fmt.Sprintf("%s >>%s 2>&1", containerScriptPath, containerLogPath)}

	// set timeout
	c.Timeout = d.Timeout
	c.ContainerInstance = new(Container)
	c.ContainerInstance.Context = context.Background()
	return c, &state, nil
}
