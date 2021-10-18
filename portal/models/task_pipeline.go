package models

import (
	"bytes"
	"database/sql/driver"

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

type PipelineJob struct {
	Image     string         `json:"image" yaml:"image"`
	Steps     []PipelineStep `json:"steps" yaml:"steps"`
	OnCreate  []PipelineStep `json:"onCreate" yaml:"onCreate"`
	OnSuccess []PipelineStep `json:"onSuccess" yaml:"onSuccess"`
	OnFail    []PipelineStep `json:"onFail" yaml:"onFail"`
}

type PipelineJobWithType struct {
	PipelineJob
	Type string `json:"type"`
}

type PipelineStep struct {
	Type string   `json:"type" yaml:"type" gorm:"size:32;not null"`
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

var defaultPipeline Pipeline

func DefaultPipeline() Pipeline {
	return defaultPipeline
}

func init() {
	buffer := bytes.NewBufferString(pipelineV0dot3)
	if err := yaml.NewDecoder(buffer).Decode(&defaultPipeline); err != nil {
		panic(err)
	}
}
