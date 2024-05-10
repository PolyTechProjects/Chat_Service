package redis

import (
	"log/slog"
	"os"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
)

var RedisClient *redis.Client

func InitRedis() {
	err := godotenv.Load()
	if err != nil {
		panic(err.Error())
	}
	redisDb, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		panic(err.Error())
	}
	redisPass := os.Getenv("REDIS_PASSWORD")
	redisAddr := os.Getenv("REDIS_ADDR")
	options := &redis.Options{
		Password: redisPass,
		Addr:     redisAddr,
		DB:       redisDb,
	}
	RedisClient = redis.NewClient(options)
	slog.Info("Connected to Redis")
}

func Close() {
	slog.Info("Disconneting from Redis")
	RedisClient.Close()
}
