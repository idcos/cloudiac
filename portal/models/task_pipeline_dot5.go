// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package models

import (
	"bytes"
	"cloudiac/common"
	"database/sql/driver"
	"fmt"

	"gopkg.in/yaml.v2"
)

const pipelineV0dot5 = `
version: 0.5

plan:
  steps:
    checkout:
      type: checkout
      name: Checkout Code

    terraformInit:
      type: terraformInit
      name: Terraform Init

    terraformPlan:
      type: terraformPlan
      name: Terraform Plan

    envScan:
      type: opaScan
      name: OPA Scan

apply:
  steps:
    checkout:
      type: checkout
      name: Checkout Code

    terraformInit:
      type: terraformInit
      name: Terraform Init

    terraformPlan:
      type: terraformPlan
      name: Terraform Plan

    envScan:
      type: envScan
      name: OPA Scan

    terraformApply:
      type: terraformApply
      name: Terraform Apply

    ansiblePlay:
      type: ansiblePlay
      name: Run playbook

destroy:
  steps:
    checkout:
      type: checkout
      name: Checkout Code

    terraformInit:
      type: terraformInit
      name: Terraform Init

    terraformPlan:
      type: terraformPlan
      name: Terraform Plan

    envScan:
      type: envScan
      name: OPA Scan

    terraformDestroy:
      type: terraformDestroy
      name: Terraform Apply
`

type PipelineDot5 struct {
	Pipeline
	Plan    PipelineDot5Task `json:"plan" yaml:"plan"`
	Apply   PipelineDot5Task `json:"apply" yaml:"apply"`
	Destroy PipelineDot5Task `json:"destroy" yaml:"destroy"`
}

type PipelineDot5Task struct {
	Steps map[string]PipelineStep `json:"steps,omitempty" yaml:"steps"`

	OnSuccess *PipelineStep `json:"onSuccess,omitempty" yaml:"onSuccess"`
	OnFail    *PipelineStep `json:"onFail,omitempty" yaml:"onFail"`
}

func (p PipelineDot5) GetTask(typ string) interface{} {
	switch typ {
	case common.TaskJobPlan:
		return p.Plan
	case common.TaskJobApply:
		return p.Apply
	case common.TaskJobDestroy:
		return p.Destroy
	default:
		panic(fmt.Errorf("unknown pipeline job type '%s'", typ))
	}
}

func (v PipelineDot5) Value() (driver.Value, error) {
	return MarshalValue(v)
}

func (v *PipelineDot5) Scan(value interface{}) error {
	return UnmarshalValue(value, v)
}

var planTaskStepNames = []string{common.TaskStepCheckout, common.TaskStepTfInit, common.TaskStepTfPlan, common.TaskStepEnvScan}
var applyTaskStepNames = []string{common.TaskStepCheckout, common.TaskStepTfInit, common.TaskStepTfPlan, common.TaskStepEnvScan, common.TaskStepTfApply, common.TaskStepAnsiblePlay}
var destroyTaskStepNames = []string{common.TaskStepCheckout, common.TaskStepTfInit, common.TaskStepTfPlan, common.TaskStepEnvScan, common.TaskStepTfDestroy}

func NewPipelineDot5(content string) (PipelineDot5, error) {
	buffer := bytes.NewBufferString(content)
	pipeline := PipelineDot5{}
	err := yaml.NewDecoder(buffer).Decode(&pipeline)
	return pipeline, err
}
