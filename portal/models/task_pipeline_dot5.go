// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package models

import (
	"bytes"
	"cloudiac/common"
	"cloudiac/utils/logs"

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

# scan 和 parse 暂不开放自定义工作流
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

type PipelineDot5 struct {
	Version string           `json:"version" yaml:"version"`
	Plan    PipelineDot5Task `json:"plan" yaml:"plan"`
	Apply   PipelineDot5Task `json:"apply" yaml:"apply"`
	Destroy PipelineDot5Task `json:"destroy" yaml:"destroy"`

	EnvScan  PipelineTask `json:"envScan" yaml:"envScan"`
	EnvParse PipelineTask `json:"envParse" yaml:"envParse"`
	TplScan  PipelineTask `json:"tplScan" yaml:"tplScan"`
	TplParse PipelineTask `json:"tplParse" yaml:"tplParse"`
}

type PipelineDot5Task struct {
	Steps map[string]PipelineStep `json:"steps,omitempty" yaml:"steps"`

	OnSuccess *PipelineStep `json:"onSuccess,omitempty" yaml:"onSuccess"`
	OnFail    *PipelineStep `json:"onFail,omitempty" yaml:"onFail"`
}

var planTaskStepNames = []string{common.TaskStepCheckout, common.TaskStepTfInit, common.TaskStepTfPlan,
	common.TaskStepEnvScan}
var applyTaskStepNames = []string{common.TaskStepCheckout, common.TaskStepTfInit, common.TaskStepTfPlan,
	common.TaskStepEnvScan, common.TaskStepTfApply, common.TaskStepAnsiblePlay}
var destroyTaskStepNames = []string{common.TaskStepCheckout, common.TaskStepTfInit, common.TaskStepTfPlan,
	common.TaskStepEnvScan, common.TaskStepTfDestroy}

func decodePipelineDot5(buf *bytes.Buffer) (PipelineDot5, error) {
	p := PipelineDot5{}
	if err := yaml.NewDecoder(buf).Decode(&p); err != nil {
		return p, err
	}

	return p, nil
}

// ConvertPipelineDot5Compatibility 转换 v0.5 为之前兼容的版本
func ConvertPipelineDot5Compatibility(buf *bytes.Buffer) (Pipeline, error) {
	p := Pipeline{}
	pDot5, err := decodePipelineDot5(buf)
	if err != nil {
		return p, err
	}

	// 兼容的部分
	p.Version = pDot5.Version
	p.EnvScan = pDot5.EnvScan
	p.EnvParse = pDot5.EnvParse
	p.TplScan = pDot5.TplScan
	p.TplParse = pDot5.TplParse

	// plan
	p.Plan = pipelineV0dot5Downgrade(pDot5.Plan, planTaskStepNames)

	// apply
	p.Apply = pipelineV0dot5Downgrade(pDot5.Apply, applyTaskStepNames)

	// destroy
	p.Destroy = pipelineV0dot5Downgrade(pDot5.Destroy, destroyTaskStepNames)

	return p, nil
}

func pipelineV0dot5Downgrade(pDot5 PipelineDot5Task, taskNames []string) PipelineTask {
	p := PipelineTask{}
	p.OnSuccess = pDot5.OnSuccess
	p.OnFail = pDot5.OnFail

	p.Steps = make([]PipelineStep, 0)
	steps := pDot5.Steps
	for _, stepName := range taskNames {
		if _, ok := steps[stepName]; !ok {
			logs.Get().Warnf("step %s is omitted.", stepName)
			continue
		}
		p.Steps = append(p.Steps, steps[stepName])
	}

	return p
}
