package env

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// loads specific .env file from ENV_FILE=file variable
// by default if no ENV_FILE is specified it will just .env
func LoadENVFromEnv() {
	file := os.Getenv("ENV_FILE")
	// file might also be ..env because when using dockermgr ENV_FILE=.${ENV_NAME}.env
	// where ENV_NAME might be an empty string
	if file == "" || file == "..env" {
		// if no name is specified we will load just .env
		err := godotenv.Load(".env")
		if err != nil {
			log.Fatalf("Error could not load %s %s\n", file, err)
		}

	} else {
		err := godotenv.Load(file)
		if err != nil {
			log.Fatalf("Error could not load %s %s\n", file, err)
		}
	}

}
