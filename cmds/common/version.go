package common

import (
	"fmt"
	"cloudiac/consts"
	"os"
)

type OptionVersion struct {
	Version bool `long:"version" description:"show version"`
}

func ShowVersionIf(show bool) {
	if show {
		fmt.Printf("version: %s build: %s\n", consts.VERSION, consts.BUILD)
		os.Exit(0)
	}
}
