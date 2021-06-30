package common

import (
	"github.com/joho/godotenv"
	"log"
	"os"
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
