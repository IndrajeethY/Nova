package db

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
)

var (
	RDb *redis.Client
	ctx = context.Background()
)

func InitDB(redisURL string) (*redis.Client, error) {
	options, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}

	options.PoolSize = 10
	options.MinIdleConns = 5
	options.PoolTimeout = 30 * time.Second

	RDb = redis.NewClient(options)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err = RDb.Ping(ctx).Result(); err != nil {
		return nil, err
	}

	log.Println("Connected to Redis")
	return RDb, nil
}

func Get(key string) string {
	if RDb == nil {
		return ""
	}
	val, _ := RDb.Get(ctx, key).Result()
	return val
}

func Set(key, value string) error {
	if RDb == nil {
		return nil
	}
	return RDb.Set(ctx, key, value, 0).Err()
}

func Del(key string) error {
	if RDb == nil {
		return nil
	}
	return RDb.Del(ctx, key).Err()
}

func Exists(key string) bool {
	if RDb == nil {
		return false
	}
	val, _ := RDb.Exists(ctx, key).Result()
	return val > 0
}

func Keys(pattern string) ([]string, error) {
	if RDb == nil {
		return nil, nil
	}
	return RDb.Keys(ctx, pattern).Result()
}

func SAdd(key string, members ...interface{}) error {
	if RDb == nil {
		return nil
	}
	return RDb.SAdd(ctx, key, members...).Err()
}

func SRem(key string, members ...interface{}) error {
	if RDb == nil {
		return nil
	}
	return RDb.SRem(ctx, key, members...).Err()
}

func SMembers(key string) ([]string, error) {
	if RDb == nil {
		return nil, nil
	}
	return RDb.SMembers(ctx, key).Result()
}

func SIsMember(key string, member interface{}) bool {
	if RDb == nil {
		return false
	}
	return RDb.SIsMember(ctx, key, member).Val()
}

func FlushAll() error {
	if RDb == nil {
		return nil
	}
	return RDb.FlushAll(ctx).Err()
}

func Close() error {
	if RDb == nil {
		return nil
	}
	return RDb.Close()
}
