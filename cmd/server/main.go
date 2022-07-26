package main

import (
	"fmt"
	"os"
	"time"

	"github.com/dannyvidal/blah/pkg/env"
)

const (
	APP_NAME = "APP_NAME"
	AUTH     = "AUTH"
)

func main() {
	env.LoadENVFromEnv()

	switch os.Getenv(APP_NAME) {

	case AUTH:
		for {
			fmt.Printf("Loaded %s\n", os.Getenv("ENV_FILE"))
			time.Sleep(time.Second * 2)
		}
	default:
		for {
			fmt.Printf("Loaded Default .env\n")
			time.Sleep(time.Second * 2)

		}
	}

}
