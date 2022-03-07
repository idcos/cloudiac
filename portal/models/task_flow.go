// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package models

import (
	"bytes"
	"cloudiac/common"
	"database/sql/driver"
	"fmt"

	"gopkg.in/yaml.v2"
)

type TaskFlows struct {
	Version  string   `json:"version" yaml:"version"`
	Plan     TaskFlow `json:"plan" yaml:"plan"`
	Apply    TaskFlow `json:"apply" yaml:"apply"`
	Destroy  TaskFlow `json:"destroy" yaml:"destroy"`
	Scan     TaskFlow `json:"scan" yaml:"scan"`
	Parse    TaskFlow `json:"parse" yaml:"parse"`
	EnvScan  TaskFlow `json:"envScan" yaml:"envScan"`
	EnvParse TaskFlow `json:"envParse" yaml:"envParse"`
	TplScan  TaskFlow `json:"tplScan" yaml:"tplScan"`
	TplParse TaskFlow `json:"tplParse" yaml:"tplParse"`
}

type TaskFlow struct {
	Steps []PipelineStep `json:"steps" yaml:"steps"`
}

func (v TaskFlow) Value() (driver.Value, error) {
	return MarshalValue(v)
}

func (v *TaskFlow) Scan(value interface{}) error {
	return UnmarshalValue(value, v)
}

const taskFlowsWithScanContent = `
version: 0.2
plan:
  steps:
    - type: init
    - type: tfscan
    - type: plan

apply:
  steps:
    - type: init
    - type: tfscan
    - type: plan
    - type: apply
    - type: play

destroy:
  steps:
    - type: init
    - type: plan
      args: ["-destroy"]
    - type: destroy

scan:
  steps:
    - type: scaninit
    - type: tfscan

parse:
  steps:
    - type: scaninit
    - type: tfparse
`

const defaultTaskFlowsContent = taskFlowsWithScanContent

var defaultTaskFlows TaskFlows

func GetTaskFlow(flows *TaskFlows, typ string) (TaskFlow, error) {
	switch typ {
	case common.TaskTypePlan:
		return flows.Plan, nil
	case common.TaskTypeApply:
		return flows.Apply, nil
	case common.TaskTypeDestroy:
		return flows.Destroy, nil
	case common.TaskTypeScan:
		return flows.Scan, nil
	case common.TaskTypeParse:
		return flows.Parse, nil
	case common.TaskTypeEnvScan:
		return flows.EnvScan, nil
	case common.TaskTypeEnvParse:
		return flows.EnvParse, nil
	case common.TaskTypeTplScan:
		return flows.TplScan, nil
	case common.TaskTypeTplParse:
		return flows.TplParse, nil
	default:
		return TaskFlow{}, fmt.Errorf("unknown task type: %v", typ)
	}
}

func DefaultTaskFlow(typ string) (TaskFlow, error) {
	return GetTaskFlow(&defaultTaskFlows, typ)
}

func DefaultTaskFlows(version string) TaskFlows {
	return defaultTaskFlows
}

func decodeTaskFlow(taskFlowContent string) TaskFlows { //nolint:unused
	taskFlows := TaskFlows{}
	buffer := bytes.NewBufferString(taskFlowContent)
	if err := yaml.NewDecoder(buffer).Decode(&taskFlows); err != nil {
		panic(err)
	}
	return taskFlows
}

func init() {
	defaultTaskFlows = decodeTaskFlow(defaultTaskFlowsContent)
}
