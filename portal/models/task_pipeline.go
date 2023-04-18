// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package models

import (
	"bytes"
	"cloudiac/portal/consts/e"
	"database/sql/driver"
	"fmt"

	"gopkg.in/yaml.v2"
)

// 通用 pipeline 接口
type IPipeline interface {
	GetVersion() string
	GetTaskFlowWithPipeline(string) PipelineTaskFlow
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

type PipelineTaskFlow struct {
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

func (v PipelineTaskFlow) Value() (driver.Value, error) {
	return MarshalValue(v)
}

func (v *PipelineTaskFlow) Scan(value interface{}) error {
	return UnmarshalValue(value, v)
}

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
			err = e.New(e.InvalidPipelineVersion)
		}

		if err != nil {
			panic(err)
		}
		defaultPipelines[v] = p
	}
}
