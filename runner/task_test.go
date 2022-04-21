package runner

import (
	"cloudiac/configs"
	"cloudiac/utils/logs"
	"testing"
)

func TestUp2Workspace(t *testing.T) {
	task := Task{
		req:       RunTaskReq{},
		logger:    logs.Get(),
		config:    configs.RunnerConfig{},
		workspace: "",
	}

	cases := []struct {
		workdir string
		except  string
	}{
		{"subdir1", "../../_cloudiac.tf"},
		{"subdir1/subdir2", "../../../_cloudiac.tf"},
		{"subdir1/subdir2/subdir3", "../../../../_cloudiac.tf"},
	}

	for _, c := range cases {
		task.req.Env.Workdir = c.workdir
		path := task.up2Workspace("_cloudiac.tf")
		if path != c.except {
			t.Errorf("task.up2Workspace, workdir %s, got %s, except %s", c.workdir, path, c.except)
			t.Fail()
		}
	}
}
