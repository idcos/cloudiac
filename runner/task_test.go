package runner

import (
	"cloudiac/configs"
	"cloudiac/utils/logs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUp2Workspace(t *testing.T) {
	task := Task{
		req:       RunTaskReq{},
		logger:    logs.Get(),
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

func TestGenTerraformrcFile(t *testing.T) {
	configs.Set(&configs.Config{})

	task := Task{
		req:       RunTaskReq{},
		logger:    logs.Get(),
		workspace: "",
	}

	mirrorUrl := "https://registry.example.org/v1/mirrors/providers/"
	cases := []struct {
		offline bool
		mirror  string
		except  string
	}{
		{false, mirrorUrl, `provider_installation {
  filesystem_mirror {
    path = "/cloudiac/terraform/plugins"
  }
  
  
  network_mirror {
	url = "https://registry.example.org/v1/mirrors/providers/"
	include = ["registry.terraform.io/*/*"]
	exclude=["registry.terraform.io/idcos/*"]
  }


  direct {
	exclude = [
	  "registry.terraform.io/*/*"
	]
  }
}`},
		{true, mirrorUrl, `provider_installation {
  filesystem_mirror {
	path = "/cloudiac/terraform/plugins"
  }


  network_mirror {
	url = "https://registry.example.org/v1/mirrors/providers/"
	include = ["registry.terraform.io/*/*"]
	exclude=["registry.terraform.io/idcos/*"]
  }


  direct {
	exclude = [
	  "registry.terraform.io/*/*"
	]
  }
}`},
		{false, "", `provider_installation {
  filesystem_mirror {
	path = "/cloudiac/terraform/plugins"
  }


  direct {
	exclude = [
	  "registry.terraform.io/idcos/*"
	]
  }
}`},
		{true, "", `provider_installation {
  filesystem_mirror {
    path = "/cloudiac/terraform/plugins"
  }
  

  direct {
    exclude = [
      "registry.terraform.io/*/*"
    ]
  }
}`},
	}

	dir, err := os.MkdirTemp("", "cloudiac-runner-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	for _, c := range cases {
		task.req.NetworkMirror = c.mirror
		configs.Get().Runner.OfflineMode = c.offline
		if err := task.genTerraformrcFile(dir); err != nil {
			t.Logf("genTerraformrcFile error: %v", err)
			t.FailNow()
		}

		path := filepath.Join(dir, TerraformrcFileName)
		content, err := os.ReadFile(path)
		if err != nil {
			t.Logf("readfile error: %v", err)
			t.FailNow()
		}

		if !assert.Equal(t, removeSpace(c.except), removeSpace(string(content)),
			"task.genTerraformrcFile, offline=%v, mirrorUrl=%s\ngot:\n%s\n except:\n%s\n",
			c.offline, c.mirror, content, c.except,
		) {
			t.FailNow()
		}
	}
}

func removeSpace(s string) string {
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "\t", "")
	s = strings.ReplaceAll(s, "\n", "")
	return s
}
