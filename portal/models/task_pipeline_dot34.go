// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package models

import (
	"bytes"
	"cloudiac/common"
	"database/sql/driver"
	"fmt"

	"gopkg.in/yaml.v2"
)

// pipeline 0.3 和 0.4 版本
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

func NewPipelineDot34(content string) (PipelineDot34, error) {
	buffer := bytes.NewBufferString(content)
	pipeline := PipelineDot34{}
	err := yaml.NewDecoder(buffer).Decode(&pipeline)
	return pipeline, err
}

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

func (p PipelineDot34) GetTask(typ string) PipelineTaskDot34 {
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

func (p PipelineDot34) GetTaskFlowWithPipeline(typ string) PipelineTaskFlow {
	task := p.GetTask(typ)

	return PipelineTaskFlow(task)
}

func (p PipelineDot34) GetVersion() string {
	return p.Version
}

func (v PipelineTaskDot34) Value() (driver.Value, error) {
	return MarshalValue(v)
}

func (v *PipelineTaskDot34) Scan(value interface{}) error {
	return UnmarshalValue(value, v)
}
