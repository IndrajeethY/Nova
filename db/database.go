package db

import (
	"context"

	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
)

var (
	RDb *redis.Client
)

func InitDB(redisURL string) (db *redis.Client, err error) {
	ctx := context.Background()
	options, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Println("Error parsing Redis URL:", err)
		return nil, err
	}

	RDb = redis.NewClient(options)

	_, err = RDb.Ping(ctx).Result()
	if err != nil {
		log.Println("Error connecting to Redis:", err)
		return nil, err
	}
	return RDb, nil
}
