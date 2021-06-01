package runner

import (
	"cloudiac/configs"
	"cloudiac/utils/logs"
	"html/template"
	"os"
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

type State struct {
	Address string
	Scheme  string
	Path    string
}

func GenStateFile(address string, scheme string, path string, targetPath string, saveState bool) {
	log := logs.Get()
	state := new(State)
	state.Address = address
	if scheme != "" {
		state.Scheme = scheme
	}
	state.Path = path
	targetFile, err := os.OpenFile(targetPath+"/state.tf", os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		log.Error("open failed err:", err)
		return
	}

	t, err := template.ParseFiles(configs.Get().Runner.AssetPath + "/state.tf.tmpl")
	if err != nil {
		log.Error("open template file err:", err)
		return
	}
	t.Execute(targetFile, state)
}
