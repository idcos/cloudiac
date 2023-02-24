// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.
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
	Mode         bool `long:"mode" short:"m" description:"update resource mode" required:"false"`
}

func (*UpdateDb) Usage() string {
	return `<updateDB resource -d>`
}

func (date *UpdateDb) Execute(args []string) error {
	dbInit()
	if err := UpdateResourceData(date); err != nil {
		fmt.Fprintf(os.Stderr, "update resource error: %s\n", err.Error())
		return err
	}
	return nil
}

func dbInit() {
	configs.Init(opt.Config)
	fmt.Println("config init finished")

	db.Init(configs.Get().Mysql)
	fmt.Println("mysql init finished")

	models.Init(false)
}

func UpdateResourceData(data *UpdateDb) error {
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
			if data.Resource.Dependencies {
				addDependenciesByStateJSON(tfstateRecords[left:right])
			}
			if data.Resource.Mode {
				UpdateModeByStateJSON(tfstateRecords[left:right])
			}
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

func UpdateModeByStateJSON(tfstateRecords []models.DBStorage) {
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
			if _, err := sess.Model(&models.Resource{}).
				Where("task_id = ? and address = ?", taskId, r.Address).
				UpdateColumn("mode", r.Mode); err != nil {
				fmt.Fprintf(os.Stderr,
					"update mode error: %s, task_id: %s, address: %s, index: %d, mode: %+v\n",
					err.Error(), taskId, r.Address, index, r.Mode)
				return
			}
		}
	}
}
