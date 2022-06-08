// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package models

import (
	"bytes"
	"cloudiac/common"
	"database/sql/driver"
	"errors"
	"fmt"

	"gopkg.in/yaml.v2"
)

// 通用 pipeline 接口
type IPipeline interface {
	GetTask(typ string) interface{}
}

type Pipeline struct {
	Version string `json:"version" yaml:"version"`
}

func (p *Pipeline) GetVersion(content string) (string, error) {
	if p.Version != "" {
		return p.Version, nil
	}

	buffer := bytes.NewBufferString(content)
	err := yaml.NewDecoder(buffer).Decode(&p)

	return p.Version, err
}

// pipeline 0.3 和 0.4 版本
type PipelineDot34 struct {
	Pipeline
	Plan    PipelineTaskDot34 `json:"plan" yaml:"plan"`
	Apply   PipelineTaskDot34 `json:"apply" yaml:"apply"`
	Destroy PipelineTaskDot34 `json:"destroy" yaml:"destroy"`

	// 0.3 pipeline 扫描步骤
	PolicyScan  PipelineTaskDot34 `json:"scan" yaml:"scan"`
	PolicyParse PipelineTaskDot34 `json:"parse" yaml:"parse"`

	// 0.4 pipeline 扫描步骤
	EnvScan  PipelineTaskDot34 `json:"envScan" yaml:"envScan"`
	EnvParse PipelineTaskDot34 `json:"envParse" yaml:"envParse"`
	TplScan  PipelineTaskDot34 `json:"tplScan" yaml:"tplScan"`
	TplParse PipelineTaskDot34 `json:"tplParse" yaml:"tplParse"`
}

type PipelineTaskDot34 struct {
	Image string         `json:"image,omitempty" yaml:"image"`
	Steps []PipelineStep `json:"steps,omitempty" yaml:"steps"`

	OnSuccess *PipelineStep `json:"onSuccess,omitempty" yaml:"onSuccess"`
	OnFail    *PipelineStep `json:"onFail,omitempty" yaml:"onFail"`
}

type PipelineStep struct {
	Type       string   `json:"type,omitempty" yaml:"type" gorm:"size:32;not null"`
	Name       string   `json:"name,omitempty" yaml:"name" gorm:"size:32;not null"`
	Timeout    int      `json:"timeout,omitempty" yaml:"timeout"`
	BeforeCmds StrSlice `json:"before,omitempty" yaml:"before" gorm:"type:text"`
	AfterCmds  StrSlice `json:"after,omitempty" yaml:"after" gorm:"type:text"`
	Args       StrSlice `json:"args,omitempty" yaml:"args" gorm:"type:text"`
}

func NewPipelineDot34(content string) (PipelineDot34, error) {
	buffer := bytes.NewBufferString(content)
	pipeline := PipelineDot34{}
	err := yaml.NewDecoder(buffer).Decode(&pipeline)
	return pipeline, err
}

func (p PipelineDot34) GetTask(typ string) interface{} {
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

func (v PipelineTaskDot34) Value() (driver.Value, error) {
	return MarshalValue(v)
}

func (v *PipelineTaskDot34) Scan(value interface{}) error {
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
		"0.5": pipelineV0dot5,
	}
	defaultPipelines = make(map[string]IPipeline)
)

func DefaultPipelineRaw() string {
	return defaultPipelineTpls[DefaultPipelineVersion]
}

func DefaultPipeline() IPipeline {
	return MustGetPipelineByVersion(DefaultPipelineVersion)
}

func GetPipelineByVersion(version string) (IPipeline, bool) {
	p, ok := defaultPipelines[version]
	return p, ok
}

func MustGetPipelineByVersion(version string) IPipeline {
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
		var p IPipeline
		var err error

		switch v {
		case "0.3", "0.4":
			p, err = NewPipelineDot34(tpl)
		case "0.5":
			p, err = NewPipelineDot5(tpl)
		default:
			err = errors.New("wrong pipeline version")
		}

		if err != nil {
			panic(err)
		}
		defaultPipelines[v] = p
	}
}
