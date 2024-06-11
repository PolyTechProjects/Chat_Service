package redis

import (
	"context"
	"fmt"
	"log/slog"

	"example.com/notification/src/config"
	"github.com/go-redis/redis/v8"
)

var RedisClient *redis.Client

func Init(cfg *config.Config) {
	redisAddr := fmt.Sprintf("%v:%v", cfg.Redis.Host, cfg.Redis.InnerPort)
	options := &redis.Options{
		Password: cfg.Redis.Password,
		Addr:     redisAddr,
		DB:       cfg.Redis.Db,
	}
	RedisClient = redis.NewClient(options)
	_, err := RedisClient.Ping(context.Background()).Result()
	if err != nil {
		panic(err.Error())
	}
	slog.Info("Connected to Redis")
}

func Close() {
	slog.Info("Disconneting from Redis")
	RedisClient.Close()
}
