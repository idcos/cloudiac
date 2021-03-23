package runner

import (
	"fmt"
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
	Schema  string
	Path    string
}

func GenStateFile(address string, scheme string, path string, targetPath string) {
	pwd, _ := os.Getwd()
	fmt.Println("open failed err:", pwd)
	state := new(State)
	state.Address = address
	state.Schema = scheme
	state.Path = path
	targetFile, err := os.OpenFile(targetPath+"/state.tf", os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		fmt.Println("open failed err:", err)
		return
	}
	t, err := template.ParseFiles("./state.tf.tmpl")
	if err != nil {
		fmt.Println("open template file err:", err)
		return
	}
	t.Execute(targetFile, state)
}
