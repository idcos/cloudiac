package runner

import (
	"bytes"
	"cloudiac/configs"
	"cloudiac/utils/logs"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

type IaCTemplate struct {
	TemplateUUID string
	TaskId       string
}

type StateStore struct {
	SaveState           bool   `json:"save_state"`
	Backend             string `json:"backend" default:"consul"`
	Scheme              string `json:"scheme" default:"http"`
	StateKey            string `json:"state_key"`
	StateBackendAddress string `json:"state_backend_address"`
	Lock                bool   `json:"lock" defalt:"true"`
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
}

type CommitedTask struct {
	TemplateId       string `json:"templateId"`
	TaskId           string `json:"taskId"`
	ContainerId      string `json:"containerId"`
	LogContentOffset int    `json:"offset"`
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

// 从指定位置读取日志文件
func ReadLogFile(filepath string, offset int, maxLines int) ([]string, error) {
	var lines []string
	// TODO(ZhengYue): 优化文件读取，考虑使用seek跳过偏移行数
	file, err := ioutil.ReadFile(filepath)
	if err != nil {
		return lines, err
	}
	buf := bytes.NewBuffer(file)
	lineCount := 0
	for {
		line, err := buf.ReadString('\n')
		if len(line) == 0 {
			if err != nil {
				if err == io.EOF {
					break
				}
				return lines, err
			}
		}
		lineCount += 1
		if lineCount > offset {
			// 未达到偏移位置，继续读取
			lines = append(lines, line)
		}
		if len(lines) == maxLines {
			// 达到最大行数，立即返回
			return lines, err
		}
		if err != nil && err != io.EOF {
			return lines, err
		}
	}
	return lines, nil
}

func GetTemplateTaskPath(templateUUID string, taskId string) string {
	conf := configs.Get()
	templateDir := fmt.Sprintf("%s/%s/%s", conf.Runner.LogBasePath, templateUUID, taskId)
	return templateDir
}

func FetchTaskLog(templateUUID string, taskId string) ([]byte, error) {
	conf := configs.Get()
	templateDir := fmt.Sprintf("%s/%s/%s", conf.Runner.LogBasePath, templateUUID, taskId)
	logFile := fmt.Sprintf("%s/%s", templateDir, ContainerLogFileName)
	return ioutil.ReadFile(logFile)
}

func CreateTemplatePath(templateUUID string, taskId string) (string, error) {
	conf := configs.Get()
	templateDir := fmt.Sprintf("%s/%s/%s", conf.Runner.LogBasePath, templateUUID, taskId)
	err := PathCreate(templateDir)
	return templateDir, err
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
func ReqToCommand(req *http.Request) (*Command, *StateStore, *IaCTemplate, error) {
	bodyData, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, nil, nil, err
	}

	logger := logs.Get()
	logger.Tracef("new task request: %s", bodyData)
	var d ReqBody
	if err := json.Unmarshal(bodyData, &d); err != nil {
		return nil, nil, nil, err
	}

	c := new(Command)

	//state := new(StateStore)
	state := d.StateStore
	// state.SaveState = d.StateStore.SaveState
	// state.Backend = d.StateStore.Backend
	// state.StateBackendAddress = d.StateStore.StateBackendAddress
	// state.StateKey = d.StateStore.StateKey
	iaCTemplate := &IaCTemplate{
		TemplateUUID: d.TemplateUUID,
		TaskId:       d.TaskID,
	}

	if d.DockerImage == "" {
		conf := configs.Get()
		c.Image = conf.Runner.DefaultImage
	} else {
		c.Image = d.DockerImage
	}

	env := []string{
		ContainerEnvTerraform,
	}
	for k, v := range d.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	c.Env = env

	// TODO(ZhengYue): 优化命令组装方式
	var cmdList []string
	logCmd := fmt.Sprintf(">> %s%s 2>&1 ", ContainerLogFilePath, ContainerLogFileName)
	ansibleCmd := fmt.Sprint(" if [ -e run.sh ];then chmod +x run.sh && ./run.sh;fi")

	// FIXME: FOR DEBUG
	cmdList = append(cmdList, fmt.Sprintf("for I in `seq 1 30`; do date && sleep 1; done %s &&", logCmd))

	cmdList = append(cmdList, fmt.Sprintf("git clone %s %s &&", d.Repo, logCmd))
	// get folder name
	s := strings.Split(d.Repo, "/")
	f := s[len(s)-1]
	f = f[:len(f)-4]

	cmdList = append(cmdList, fmt.Sprintf("cd %s %s &&", f, logCmd))
	cmdList = append(cmdList, fmt.Sprintf("git checkout  %s %s &&", d.RepoBranch, logCmd))
	cmdList = append(cmdList, fmt.Sprintf("git checkout -b run_branch %s %s &&", d.RepoCommit, logCmd))
	cmdList = append(cmdList, fmt.Sprintf("cp %sstate.tf . &&", ContainerLogFilePath))
	cmdList = append(cmdList, fmt.Sprintf("terraform init  -plugin-dir %s %s &&", ContainerProviderPath, logCmd))
	if d.Mode == "apply" {
		log.Println("entering apply mode ...")
		if d.Varfile != "" {
			cmdList = append(cmdList, fmt.Sprintf("%s %s %s &&%s %s &&%s %s",
				"terraform apply -auto-approve -var-file ", d.Varfile, logCmd, ansibleCmd, logCmd, d.Extra, logCmd))
		} else {
			cmdList = append(cmdList, fmt.Sprintf("%s %s &&%s %s &&%s %s", "terraform apply -auto-approve ", logCmd, ansibleCmd, logCmd, d.Extra, logCmd))
		}

	} else if d.Mode == "destroy" {
		log.Println("entering destroy mode ...")
		cmdList = append(cmdList, fmt.Sprintf("%s %s&&%s", "terraform destroy -auto-approve -var-file", d.Varfile, d.Extra))
	} else if d.Mode == "pull" {
		log.Println("show state info ...")
		cmdList = append(cmdList, fmt.Sprintf("%s&&%s", "terraform state pull", d.Extra))
	} else {
		if d.Varfile != "" {
			cmdList = append(cmdList, fmt.Sprintf("%s %s %s", "terraform plan -var-file ", d.Varfile, logCmd))
		} else {
			cmdList = append(cmdList, fmt.Sprintf("%s %s", "terraform plan  ", logCmd))
		}

		//cmdList = append(cmdList, fmt.Sprintf("%s %s&&%s", "terraform plan -var-file", d.Varfile, d.Extra))
	}

	cmdstr := ""
	for _, v := range cmdList {
		cmdstr = cmdstr + v
	}
	var t []string
	t = append(t, "sh")
	t = append(t, "-c")
	t = append(t, cmdstr)

	c.Commands = t

	// set timeout
	c.Timeout = d.Timeout

	c.ContainerInstance = new(Container)
	c.ContainerInstance.Context = context.Background()
	log.Printf("%#v", c)
	return c, &state, iaCTemplate, nil
}

func LineCounter(r io.Reader) (int, error) {
	buf := make([]byte, 32*1024)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := r.Read(buf)
		count += bytes.Count(buf[:c], lineSep)

		switch {
		case err == io.EOF:
			return count, nil

		case err != nil:
			return count, err
		}
	}
}
