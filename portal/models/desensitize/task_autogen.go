// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

// Code generated by code-gen/desenitize DO NOT EDIT

package desensitize

import (
	// "encoding/json"
	"cloudiac/portal/models"
)

type Task struct {
	models.Task
}


// 不定义 MarshalJSON() 方法，因为一旦定义了该结构体就无法组合使用了，
// 会覆盖 MarshalJSON() 方法以导致组合的其他字段不输出。 比如定义结构体:
// type TaskWithExt struct {
// 		models.Task
//		Ext	string
// }
// 当我们调用 json.Marshal(TaskWithExt{}) 时 Ext 字段不会输出，
// 因为直接调用了 models.Task.MarshalJSON() 方法。
// func (v Task) MarshalJSON() ([]byte, error) {
// 	return json.Marshal(v.Task.Desensitize())
// }
func (v Task) Desensitize() Task {
	return Task{v.Task.Desensitize()}
}

func NewTask(v models.Task) Task {
	rv := Task{v.Desensitize()}
	return rv
}

func NewTaskPtr(v *models.Task) *Task {
	rv := Task{v.Desensitize()}
	return &rv
}

func NewTaskSlice(vs []models.Task) []Task {
	rvs := make([]Task, len(vs))
	for i := 0; i < len(vs); i++ {
		rvs[i] = NewTask(vs[i])
	}
	return rvs
}

func NewTaskSlicePtr(vs []*models.Task) []*Task {
	rvs := make([]*Task, len(vs))
	for i := 0; i < len(vs); i++ {
		v := NewTask(*vs[i])
		rvs[i] = &v
	}
	return rvs
}
