package models

import (
	"bytes"
	"cloudiac/common"
	"database/sql/driver"
	"fmt"

	"gopkg.in/yaml.v2"
)

type Pipeline struct {
	Version     string      `json:"version" yaml:"version"`
	Plan        PipelineJob `json:"plan" yaml:"plan"`
	Apply       PipelineJob `json:"apply" yaml:"apply"`
	Play        PipelineJob `json:"play" yaml:"play"`
	DestroyPlan PipelineJob `json:"destroyPlan" yaml:"destroyPlan"`
	Destroy     PipelineJob `json:"destroy" yaml:"destroy"`

	// 直接命名为 Scan 会与 Scan() 接口方法重名，所以这里命名为 PolicyScan
	PolicyScan  PipelineJob `json:"scan" yaml:"scan"`
	PolicyParse PipelineJob `json:"parse" yaml:"parse"`
}

func (p Pipeline) GetJob(typ string) PipelineJob {
	switch typ {
	case common.TaskJobPlan:
		return p.Plan
	case common.TaskJobApply:
		return p.Apply
	case common.TaskJobPlay:
		return p.Play
	case common.TaskJobDestroyPlan:
		return p.DestroyPlan
	case common.TaskJobDestroy:
		return p.Destroy
	case common.TaskJobScan:
		return p.PolicyScan
	case common.TaskJobParse:
		return p.PolicyParse
	default:
		panic(fmt.Errorf("unknown pipeline job type '%s'", typ))
	}
}

type PipelineJob struct {
	Image string         `json:"image,omitempty" yaml:"image"`
	Steps []PipelineStep `json:"steps,omitempty" yaml:"steps"`

	// 定义为指针类型，这样在字段无值时 json 序列化不会输出该字段，避免写入数据库时记录为空结构体
	OnCreate  *PipelineStep `json:"onCreate,omitempty" yaml:"onCreate"`
	OnSuccess *PipelineStep `json:"onSuccess,omitempty" yaml:"onSuccess"`
	OnFail    *PipelineStep `json:"onFail,omitempty" yaml:"onFail"`
}

type PipelineJobWithType struct {
	PipelineJob
	Type string `json:"type"`
}

type PipelineStep struct {
	Type string   `json:"type,omitempty" yaml:"type" gorm:"size:32;not null"`
	Name string   `json:"name,omitempty" yaml:"name" gorm:"size:32;not null"`
	Args StrSlice `json:"args,omitempty" yaml:"args" gorm:"type:text"`
}

func (v Pipeline) Value() (driver.Value, error) {
	return MarshalValue(v)
}

func (v *Pipeline) Scan(value interface{}) error {
	return UnmarshalValue(value, v)
}

const pipelineV0dot3 = `
version: 0.3

plan:
  steps:
    - type: checkout
      name: Checkout code
      
    - type: terraformInit
      name: Terraform Init

    - type: terraformPlan
      name: Terraform Plan

apply:
  steps:
    - type: terraformApply
      name: Terraform Apply

play:
  steps:
     - type: ansiblePlay
       name: Run playbook

destroyPlan:
  steps:
    - type: checkout
      name: Checkout code
      
    - type: terraformInit
      name: Terraform Init

    - type: terraformPlan
      name: Terraform Plan
      args: 
        - "-destroy"
      
destroy:
  steps:
    - type: terraformDestroy
      name: Terraform Destroy

scan:
  steps:
    - type: scaninit
    - type: tfscan

parse:
  steps:
    - type: scaninit
    - type: tfparse
`

const defaultPipelineVersion = "0.3"

var (
	defaultPipelineTpls = map[string]string{
		"0.3": pipelineV0dot3,
	}
	defaultPipelines = make(map[string]Pipeline)
)

func DefaultPipeline() Pipeline {
	return MustGetPipelineByVersion(defaultPipelineVersion)
}

func GetPipelineByVersion(version string) (Pipeline, bool) {
	p, ok := defaultPipelines[version]
	return p, ok
}

func MustGetPipelineByVersion(version string) Pipeline {
	pipeline, ok := GetPipelineByVersion(version)
	if !ok {
		panic(fmt.Errorf("pipeline for version '%s' not exists", version))
	}
	return pipeline
}

func init() {
	for v, tpl := range defaultPipelineTpls {
		buffer := bytes.NewBufferString(tpl)
		p := Pipeline{}
		if err := yaml.NewDecoder(buffer).Decode(&p); err != nil {
			panic(err)
		}
		defaultPipelines[v] = p
	}
}
