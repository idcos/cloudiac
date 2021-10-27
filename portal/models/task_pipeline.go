package models

import (
	"bytes"
	"cloudiac/common"
	"database/sql/driver"
	"fmt"

	"gopkg.in/yaml.v2"
)

type Pipeline struct {
	Version string       `json:"version" yaml:"version"`
	Plan    PipelineTask `json:"plan" yaml:"plan"`
	Apply   PipelineTask `json:"apply" yaml:"apply"`
	Destroy PipelineTask `json:"destroy" yaml:"destroy"`

	// 直接命名为 Scan 会与 Scan() 接口方法重名，所以这里命名为 PolicyScan
	PolicyScan  PipelineTask `json:"scan" yaml:"scan"`
	PolicyParse PipelineTask `json:"parse" yaml:"parse"`
}

func (p Pipeline) GetTask(typ string) PipelineTask {
	switch typ {
	case common.TaskJobPlan:
		return p.Plan
	case common.TaskJobApply:
		return p.Apply
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

type PipelineTask struct {
	Image string         `json:"image,omitempty" yaml:"image"`
	Steps []PipelineStep `json:"steps,omitempty" yaml:"steps"`

	OnSuccess *PipelineStep `json:"onSuccess,omitempty" yaml:"onSuccess"`
	OnFail    *PipelineStep `json:"onFail,omitempty" yaml:"onFail"`
}

type PipelineTaskWithType struct {
	PipelineTask
	Type string `json:"type"`
}

type PipelineStep struct {
	Type string   `json:"type,omitempty" yaml:"type" gorm:"size:32;not null"`
	Name string   `json:"name,omitempty" yaml:"name" gorm:"size:32;not null"`
	Args StrSlice `json:"args,omitempty" yaml:"args" gorm:"type:text"`
}

func (v PipelineTask) Value() (driver.Value, error) {
	return MarshalValue(v)
}

func (v *PipelineTask) Scan(value interface{}) error {
	return UnmarshalValue(value, v)
}

const pipelineV0dot3 = `
version: 0.3

plan:
  steps:
    - type: checkout
      name: Checkout Code
      
    - type: terraformInit
      name: Terraform Init

    - type: opaScan
      name: OPA Scan

    - type: terraformPlan
      name: Terraform Plan

apply:
  steps:
    - type: checkout
      name: Checkout Code
      
    - type: terraformInit
      name: Terraform Init

    - type: opaScan
      name: OPA Scan

    - type: terraformPlan
      name: Terraform Plan

    - type: terraformApply
      name: Terraform Apply

    - type: ansiblePlay
      name: Run playbook
      
destroy:
  steps:
    - type: checkout
      name: Checkout Code
      
    - type: terraformInit
      name: Terraform Init

    - type: terraformPlan
      name: Terraform Plan
      args: 
        - "-destroy"

    - type: terraformDestroy
      name: Terraform Destroy


# scan 和 parse 暂不开发自定义工作流
scan:
  steps:
    - type: scaninit
    - type: opaScan

parse:
  steps:
    - type: scaninit
    - type: regoParse 
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
	if version == "" {
		version = defaultPipelineVersion
	}
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
