package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

type ConfigType struct {
	ApiId         int32
	ApiHash       string
	DbUrl         string
	Token         string
	StringSession string
}

var Config *ConfigType

func LoadConfig() (*ConfigType, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	apiId, err := strconv.Atoi(os.Getenv("API_ID"))
	if err != nil {
		log.Fatalf("Error parsing API_ID: %v", err)
	}

	Config = &ConfigType{
		ApiId:         int32(apiId),
		ApiHash:       os.Getenv("API_HASH"),
		DbUrl:         os.Getenv("DB_URL"),
		Token:         os.Getenv("TOKEN"),
		StringSession: os.Getenv("STRING_SESSION"),
	}

	return Config, nil
}
