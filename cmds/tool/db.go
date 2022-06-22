// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.
package main

import (
	"cloudiac/configs"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/portal/services"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

type UpdateDb struct {
	Resource ResourceData `long:"resource" short:"r" description:"update resource data" required:"false"`
}

type ResourceData struct {
	Dependencies bool `long:"dependencies" short:"d" description:"update resource dependencies" required:"false"`
}

func (*UpdateDb) Usage() string {
	return `<updateDB resource -d>`
}

func (date *UpdateDb) Execute(args []string) error {
	dbInit()
	if date.Resource.Dependencies {
		if err := AddDependenciesData(); err != nil {
			fmt.Fprintf(os.Stderr, "add dependecies error: %s\n", err.Error())
			return err
		}
	}
	return nil
}

func dbInit() {
	configs.Init(opt.Config)
	fmt.Printf("mysql uri: %+v\n", configs.Get().Mysql)

	db.Init(configs.Get().Mysql)
	models.Init(false)
}

func AddDependenciesData() error {
	sess, t := db.Get(), time.Now()

	fmt.Println("start to search all tfstate.json")
	var tfstateRecords []models.DBStorage
	if err := sess.Model(&models.DBStorage{}).Where("path LIKE ?", "%tfstate.json").
		Select("path", "content").Find(&tfstateRecords); err != nil {
		return err
	}

	size := len(tfstateRecords)

	if size == 0 {
		return nil
	}

	fmt.Printf("found total %d tfstate.json data\n", len(tfstateRecords))

	wg, step := sync.WaitGroup{}, size/5
	for i := 0; i < size; i = i + step {
		left, right := i, i+step
		if i+step > size {
			right = size - 1
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			addDependenciesByStateJSON(tfstateRecords[left:right])
		}()
	}

	wg.Wait()

	fmt.Printf("update resource date done, timeCost: %s\n", time.Since(t).String())

	return nil
}

func addDependenciesByStateJSON(tfstateRecords []models.DBStorage) {
	sess := db.Get()
	size := len(tfstateRecords)
	for index, row := range tfstateRecords {
		fmt.Printf("process the %s, index: %d, total: %d\n", row.Path, index, size)
		tfState, err := services.UnmarshalStateJson(row.Content)
		if err != nil {
			fmt.Fprintf(os.Stderr, "index: %d, unmarshal tfstate error: %s, path: %s, content: %s\n",
				index, err, row.Path, string(row.Content))
			continue
		}

		taskId := strings.Split(row.Path, "/")[2]

		rs := make([]*models.Resource, 0)
		rs = append(rs, services.TraverseStateModule(&tfState.Values.RootModule)...)
		for i := range tfState.Values.ChildModules {
			rs = append(rs, services.TraverseStateModule(&tfState.Values.ChildModules[i])...)
		}

		for _, r := range rs {
			if r.Dependencies == nil {
				continue
			}
			if _, err := sess.Model(&models.Resource{}).
				Where("task_id = ? and address = ?", taskId, r.Address).
				UpdateColumn("dependencies", r.Dependencies); err != nil {
				fmt.Fprintf(os.Stderr,
					"update dependencies error: %s, task_id: %s, address: %s, index: %d, dependencies: %+v\n",
					err.Error(), taskId, r.Address, index, r.Dependencies)
				return
			}
		}
	}
}
