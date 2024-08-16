// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package models

import (
	"cloudiac/common"
	"cloudiac/portal/libs/db"
	"cloudiac/runner"
	"cloudiac/utils"
	"fmt"
	"path"
)

const (
	TaskStepInit     = common.TaskStepTfInit
	TaskStepPlan     = common.TaskStepTfPlan
	TaskStepApply    = common.TaskStepTfApply
	TaskStepDestroy  = common.TaskStepTfDestroy
	TaskStepPlay     = common.TaskStepAnsiblePlay
	TaskStepCommand  = common.TaskStepCommand
	TaskStepCollect  = common.TaskStepCollect
	TaskStepEnvParse = common.TaskStepEnvParse
	TaskStepEnvScan  = common.TaskStepEnvScan
	TaskStepTplParse = common.TaskStepTplParse
	TaskStepTplScan  = common.TaskStepTplScan
	TaskStepScanInit = common.TaskStepScanInit
	TaskStepOpaScan  = common.TaskStepOpaScan

	TaskStepPending   = common.TaskStepPending
	TaskStepApproving = common.TaskStepApproving
	TaskStepRejected  = common.TaskStepRejected
	TaskStepRunning   = common.TaskStepRunning
	TaskStepFailed    = common.TaskStepFailed
	TaskStepComplete  = common.TaskStepComplete
	TaskStepTimeout   = common.TaskStepTimeout
	TaskStepAborted   = common.TaskStepAborted
)

type TaskStep struct {
	BaseModel
	PipelineStep

	OrgId     Id     `json:"orgId" gorm:"size:32;not null"`
	ProjectId Id     `json:"projectId" gorm:"size:32;not null"`
	EnvId     Id     `json:"envId" gorm:"size:32;not null"`
	TaskId    Id     `json:"taskId" gorm:"index;size:32;not null"`
	NextStep  Id     `json:"nextStep" gorm:"size:32;default:''"`
	Index     int    `json:"index" gorm:"size:32;not null"`
	Status    string `json:"status" gorm:""`            // type:enum('pending','approving','rejected','running','failed','complete','timeout','aborted')
	ExitCode  int    `json:"exitCode" gorm:"default:0"` // 执行退出码，status 为 failed 时才有意义
	Message   Text   `json:"message" gorm:"type:text"`
	StartAt   *Time  `json:"startAt" gorm:""`
	EndAt     *Time  `json:"endAt" gorm:""`
	LogPath   string `json:"logPath" gorm:""`

	MustApproval bool `json:"requireApproval" gorm:""`   // 步骤需要审批
	ApproverId   Id   `json:"approverId" gorm:"size:32"` // 审批者用户 id

	CurrentRetryCount int   `json:"currentRetryCount" gorm:"size:32;default:0"` // 当前重试次数
	NextRetryTime     int64 `json:"nextRetryTime" gorm:"default:0"`             // 下次重试时间
	RetryNumber       int   `json:"retryNumber" gorm:"size:32;default:0"`       // 每个步骤可以重试的总次数

	IsCallback bool `json:"isCallback" gorm:"default:0"` // 步骤是否为回调
}

func (TaskStep) TableName() string {
	return "iac_task_step"
}

func (s TaskStep) String() string {
	return fmt.Sprintf("%s(%d)", s.Type, s.Index)
}

func (t *TaskStep) Migrate(sess *db.Session) (err error) {
	if err := sess.ModifyModelColumn(t, "type"); err != nil {
		return err
	}
	if err := sess.ModifyModelColumn(t, "status"); err != nil {
		return err
	}
	return nil
}

func (s *TaskStep) IsStarted() bool {
	return !utils.StrInArray(s.Status, TaskStepPending, TaskStepApproving)
}

func (s *TaskStep) IsExited() bool {
	return s.IsExitedStatus(s.Status)
}

func (TaskStep) IsExitedStatus(status string) bool {
	return utils.StrInArray(status,
		TaskStepRejected,
		TaskStepComplete,
		TaskStepFailed,
		TaskStepTimeout,
		TaskStepAborted)
}

// 执行成功
func (s *TaskStep) IsSuccess() bool {
	return utils.StrInArray(s.Status, TaskStepComplete)
}

// 执行失败
// 当前 IsFail() 只用于 pipeline 回调时判断任务状态，
// 在没有单独的 timeout 和 aborted 回调之前这两个状态都当作 fail
func (s *TaskStep) IsFail() bool {
	return utils.StrInArray(s.Status,
		TaskStepFailed,
		TaskStepTimeout,
		TaskStepAborted,
	)
}

func (s *TaskStep) IsApproved() bool {
	if len(s.ApproverId) == 0 {
		return false
	}
	if s.Status == TaskStepRejected {
		return false
	}
	return true
}

func (s *TaskStep) IsRejected() bool {
	return s.Status == TaskStepRejected
}

func (s *TaskStep) GenLogPath() string {
	return path.Join(
		s.ProjectId.String(),
		s.EnvId.String(),
		s.TaskId.String(),
		fmt.Sprintf("step%d", s.Index),
		runner.TaskLogName,
	)
}
