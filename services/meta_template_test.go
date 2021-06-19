package services

import (
	"cloudiac/libs/db"
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
	tx := db.Get().Begin()
	defer tx.Rollback()
	InitMetaTemplate(tx)
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
