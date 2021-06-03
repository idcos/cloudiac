package runner

import (
	"os"
	"path/filepath"
	"text/template"
)

/*
content format of state.tf file:

terraform {
  backend "consul" {
    address = "localhost:8500"
    scheme  = "http"
    path    = "tf/tfcloud-demo-repo/terraform.tfstate"
    lock    = true
    gzip    = false
  }
}
*/

const backendTemplate = `
terraform {
  backend "consul" {
    address = "{{.Address}}"
    scheme  = "{{.Scheme}}"
    path    = "{{.Path}}"
    lock    = true
    gzip    = false
  }
}`

var backendTpl = template.Must(template.New("").Parse(backendTemplate))

type State struct {
	Address string
	Scheme  string
	Path    string
}

func GenBackendConfig(address string, scheme string, path string, workingDir string) error {
	state := new(State)
	state.Address = address
	if scheme != "" {
		state.Scheme = scheme
	}
	state.Path = path
	targetFile, err := os.OpenFile(filepath.Join(workingDir, BackendConfigName), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer targetFile.Close()

	return backendTpl.Execute(targetFile, state)
}
