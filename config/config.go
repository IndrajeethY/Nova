package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
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
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
		return nil, err
	}

	apiId, err := strconv.Atoi(os.Getenv("API_ID"))
	if err != nil {
		log.Fatalf("Error parsing API_ID")
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
