package services

import (
	"cloudiac/configs"
	"cloudiac/libs/db"
	"cloudiac/utils/logs"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

func init() {
	configPath := "../config.yml"
	configs.Init(configPath)
	conf := configs.Get().Log
	db.Init()
	logs.Init(conf.LogLevel, "", 0)
}

func TestInitMetaTemplate(t *testing.T) {
	InitMetaTemplate()
}

func TestMetaAnalysis(t *testing.T) {
	f, err := os.OpenFile("../meta.yml", os.O_RDONLY, 0600)
	if err != nil {
		fmt.Println(err.Error())
	}
	defer f.Close()

	contentByte, err := ioutil.ReadAll(f)
	//fmt.Println(string(contentByte))

	mt,er:=MetaAnalysis(contentByte)
	if er!=nil{
		fmt.Println(mt)
	}
	fmt.Println(mt)
}
