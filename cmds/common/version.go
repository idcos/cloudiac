// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package common

import (
	"cloudiac/common"
	"fmt"
	"os"
)

type OptionVersion struct {
	Version bool `long:"version" description:"show version"`
}

func ShowVersionIf(show bool) {
	if show {
		fmt.Printf("version: %s build: %s\n", common.VERSION, common.BUILD)
		os.Exit(0)
	}
}

type VersionCommand struct {
}

func (VersionCommand) Execute([]string) error {
	fmt.Printf("version: %s build: %s\n", common.VERSION, common.BUILD)
	return nil
}
