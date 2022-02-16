// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

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

	// 0.3 pipeline 扫描步骤
	PolicyScan  PipelineTask `json:"scan" yaml:"scan"`
	PolicyParse PipelineTask `json:"parse" yaml:"parse"`

	// 0.4 pipeline 扫描步骤
	EnvScan  PipelineTask `json:"envScan" yaml:"envScan"`
	EnvParse PipelineTask `json:"envParse" yaml:"envParse"`
	TplScan  PipelineTask `json:"tplScan" yaml:"tplScan"`
	TplParse PipelineTask `json:"tplParse" yaml:"tplParse"`
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
	case common.TaskJobEnvScan:
		return p.EnvScan
	case common.TaskJobEnvParse:
		return p.EnvParse
	case common.TaskJobTplScan:
		return p.TplScan
	case common.TaskJobTplParse:
		return p.TplParse
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
`

const pipelineV0dot4 = `
version: 0.4

plan:
  steps:
    - type: checkout
      name: Checkout Code

    - type: terraformInit
      name: Terraform Init

    - type: terraformPlan
      name: Terraform Plan

    - type: envScan
      name: OPA Scan

apply:
  steps:
    - type: checkout
      name: Checkout Code

    - type: terraformInit
      name: Terraform Init

    - type: terraformPlan
      name: Terraform Plan

    - type: envScan
      name: OPA Scan

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
envScan:
  steps:
    - type: checkout
    - type: terraformInit
    - type: terraformPlan
    - type: envScan

envParse:
  steps:
    - type: checkout
    - type: terraformInit
    - type: terraformPlan
    - type: envParse

tplScan:
  steps:
    - type: scaninit
    - type: tplScan

tplParse:
  steps:
    - type: scaninit
    - type: tplParse
`

const DefaultPipelineVersion = "0.4"

var (
	defaultPipelineTpls = map[string]string{
		"0.3": pipelineV0dot3,
		"0.4": pipelineV0dot4,
	}
	defaultPipelines = make(map[string]Pipeline)
)

func DefaultPipelineRaw() string {
	return defaultPipelineTpls[DefaultPipelineVersion]
}

func DefaultPipeline() Pipeline {
	return MustGetPipelineByVersion(DefaultPipelineVersion)
}

func GetPipelineByVersion(version string) (Pipeline, bool) {
	p, ok := defaultPipelines[version]
	return p, ok
}

func MustGetPipelineByVersion(version string) Pipeline {
	if version == "" {
		version = DefaultPipelineVersion
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
