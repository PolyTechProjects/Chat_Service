package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Auth     AuthConfig
	Users    UsersConfig
	App      AppConfig
	Database DatabaseConfig
	Redis    RedisConfig
}

type AuthConfig struct {
	AuthHost string `env:"AUTH_HOST"`
	AuthPort string `env:"AUTH_PORT"`
}

type UsersConfig struct {
	UsersHost string `env:"USERS_HOST"`
	UsersPort string `env:"USERS_PORT"`
}

type AppConfig struct {
	HttpInnerPort int `env:"APP_HTTP_PORT"`
	GrpcInnerPort int `env:"APP_GRPC_PORT"`
}

type DatabaseConfig struct {
	DatabaseName string `env:"DB_NAME"`
	UserName     string `env:"DB_USER"`
	Password     string `env:"DB_PASSWORD"`
	Host         string `env:"DB_HOST"`
	InnerPort    int    `env:"DB_PORT"`
	SslMode      string `env:"DB_SSL_MODE"`
}

type RedisConfig struct {
	Host                    string `env:"REDIS_HOST"`
	InnerPort               int    `env:"REDIS_PORT"`
	Db                      int    `env:"REDIS_DB"`
	Password                string `env:"REDIS_PASSWORD"`
	NotificationChannelName string `env:"REDIS_NOTIFICATION_CHANNEL_NAME"`
	SubscriptionChannelName string `env:"REDIS_SUBSCRIPTION_CHANNEL_NAME"`
}

func MustLoad() *Config {
	err := godotenv.Load()
	if err != nil {
		panic("failed to load .env file: " + err.Error())
	}
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		panic("failed to read config: " + err.Error())
	}
	return &cfg
}
