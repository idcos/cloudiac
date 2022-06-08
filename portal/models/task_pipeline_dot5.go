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
      name: Checkout Code

    terraformInit:
      name: Terraform Init

    terraformPlan:
      name: Terraform Plan

    envScan:
      name: OPA Scan

apply:
  steps:
    checkout:
      name: Checkout Code

    terraformInit:
      name: Terraform Init

    terraformPlan:
      name: Terraform Plan

    envScan:
      name: OPA Scan

    terraformApply:
      name: Terraform Apply

    ansiblePlay:
      name: Run playbook

destroy:
  steps:
    checkout:
      name: Checkout Code

    terraformInit:
      name: Terraform Init

    terraformPlan:
      name: Terraform Plan
      args:
        - "-destroy"

    envScan:
      name: OPA Scan

    terraformDestroy:
      name: Terraform Apply
`

type PipelineDot5 struct {
	Pipeline
	Plan    PipelineDot5Task `json:"plan" yaml:"plan"`
	Apply   PipelineDot5Task `json:"apply" yaml:"apply"`
	Destroy PipelineDot5Task `json:"destroy" yaml:"destroy"`

	PolicyScan PipelineDot5Task `json:"scan" yaml:"scan"`
	EnvScan    PipelineDot5Task `json:"envScan" yaml:"envScan"`
}

type PipelineDot5Task struct {
	Image string                   `json:"image,omitempty" yaml:"image"`
	Steps map[string]*PipelineStep `json:"steps,omitempty" yaml:"steps"`

	OnSuccess *PipelineStep `json:"onSuccess,omitempty" yaml:"onSuccess"`
	OnFail    *PipelineStep `json:"onFail,omitempty" yaml:"onFail"`
}

func (p PipelineDot5) GetTask(typ string) PipelineDot5Task {
	switch typ {
	case common.TaskJobPlan:
		return p.Plan
	case common.TaskJobApply:
		return p.Apply
	case common.TaskJobDestroy:
		return p.Destroy
	case common.TaskJobScan:
		return p.PolicyScan
	case common.TaskJobEnvScan:
		return p.EnvScan
	default:
		panic(fmt.Errorf("unknown pipeline job type '%s'", typ))
	}
}

func (p PipelineDot5) GetTaskFlowWithPipeline(typ string) PipelineTaskFlow {
	task := p.GetTask(typ)
	return GetTaskFlowWithPipelineDot5(task, mTaskStepNames[typ])
}

func (p PipelineDot5) GetVersion() string {
	return p.Version
}

func (v PipelineDot5) Value() (driver.Value, error) {
	return MarshalValue(v)
}

func (v *PipelineDot5) Scan(value interface{}) error {
	return UnmarshalValue(value, v)
}

func GetTaskFlowWithPipelineDot5(pDot5 PipelineDot5Task, taskNames []string) PipelineTaskFlow {
	p := PipelineTaskFlow{}
	p.Image = pDot5.Image
	p.OnSuccess = pDot5.OnSuccess
	if p.OnSuccess != nil {
		p.OnSuccess.Type = common.TaskStepCommand
	}
	p.OnFail = pDot5.OnFail
	if p.OnFail != nil {
		p.OnFail.Type = common.TaskStepCommand
	}

	p.Steps = make([]PipelineStep, 0)
	steps := pDot5.Steps
	for _, stepName := range taskNames {
		steps[stepName].Type = stepName
		p.Steps = append(p.Steps, *steps[stepName])
	}

	return p
}

var mTaskStepNames = map[string][]string{
	common.TaskJobPlan:    {common.TaskStepCheckout, common.TaskStepTfInit, common.TaskStepTfPlan, common.TaskStepEnvScan},
	common.TaskJobApply:   {common.TaskStepCheckout, common.TaskStepTfInit, common.TaskStepTfPlan, common.TaskStepEnvScan, common.TaskStepTfApply, common.TaskStepAnsiblePlay},
	common.TaskJobDestroy: {common.TaskStepCheckout, common.TaskStepTfInit, common.TaskStepTfPlan, common.TaskStepEnvScan, common.TaskStepTfDestroy},
}

func NewPipelineDot5(content string) (PipelineDot5, error) {
	buffer := bytes.NewBufferString(content)
	pipeline := PipelineDot5{}
	err := yaml.NewDecoder(buffer).Decode(&pipeline)

	CompletePipelineDot5(&pipeline)
	return pipeline, err
}

func CompletePipelineDot5(p *PipelineDot5) {
	// plan
	for _, stepName := range mTaskStepNames[common.TaskJobPlan] {
		if _, ok := p.Plan.Steps[stepName]; !ok {
			p.Plan.Steps[stepName] = &PipelineStep{Name: stepName}
		}
	}

	// apply
	for _, stepName := range mTaskStepNames[common.TaskJobApply] {
		if _, ok := p.Apply.Steps[stepName]; !ok {
			p.Apply.Steps[stepName] = &PipelineStep{Name: stepName}
		}
	}

	// destroy
	for _, stepName := range mTaskStepNames[common.TaskJobDestroy] {
		if _, ok := p.Destroy.Steps[stepName]; !ok {
			p.Destroy.Steps[stepName] = &PipelineStep{Name: stepName}
		}

		// complete plan args with --destroy
		if stepName == common.TaskStepTfPlan {
			findDestroyArgs := false
			args := p.Destroy.Steps[stepName].Args
			if args == nil {
				args = []string{"-destroy"}
			}

			for _, arg := range args {
				if arg == "-destroy" {
					findDestroyArgs = true
					break
				}
			}
			if !findDestroyArgs {
				args = append(args, "-destroy")
			}

			p.Destroy.Steps[stepName].Args = args
		}
	}
}
