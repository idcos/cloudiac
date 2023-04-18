// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package common

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func LoadDotEnv() {
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			log.Panic(err)
		}
	} else if !os.IsNotExist(err) {
		log.Panic(err)
	}
}
