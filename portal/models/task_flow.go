package models

import (
	"bytes"
	"cloudiac/common"
	"database/sql/driver"
	"fmt"
	"gopkg.in/yaml.v2"
)

type TaskFlows struct {
	Version string   `json:"version" yaml:"version"`
	Plan    TaskFlow `json:"plan" yaml:"plan"`
	Apply   TaskFlow `json:"apply" yaml:"apply"`
	Destroy TaskFlow `json:"destroy" yaml:"destroy"`
}

type TaskFlow struct {
	Steps []TaskStepBody `json:"steps" yaml:"steps"`
}

func (v TaskFlow) Value() (driver.Value, error) {
	return MarshalValue(v)
}

func (v *TaskFlow) Scan(value interface{}) error {
	return UnmarshalValue(value, v)
}

const defaultTaskFlowsContent = `
version: 0.1
plan:
  steps:
    - type: init
    - type: plan

apply:
  steps:
    - type: init
    - type: plan
    - type: apply 
    - type: play
      args: ["-v"]

destroy:
  steps:
    - type: init
      args: ["-destroy"]
    - type: destroy
`

var defaultTaskFlows TaskFlows

func GetTaskFlow(flows *TaskFlows, typ string) (TaskFlow, error) {
	switch typ {
	case common.TaskTypePlan:
		return flows.Plan, nil
	case common.TaskTypeApply:
		return flows.Apply, nil
	case common.TaskTypeDestroy:
		return flows.Destroy, nil
	default:
		return TaskFlow{}, fmt.Errorf("unknown task type: %v", typ)
	}
}

func DefaultTaskFlow(typ string) (TaskFlow, error) {
	return GetTaskFlow(&defaultTaskFlows, typ)
}

func DefaultTaskFlows() TaskFlows {
	return defaultTaskFlows
}

func init() {
	buffer := bytes.NewBufferString(defaultTaskFlowsContent)
	if err := yaml.NewDecoder(buffer).Decode(&defaultTaskFlows); err != nil {
		panic(err)
	}
}
