package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	App      AppConfig
	UserMgmt UserMgmtConfig
	Db       DbConfig
}

type AppConfig struct {
	InnerGrpcPort int `env:"APP_GRPC_INNER_PORT"`
	InnerHttpPort int `env:"APP_HTTP_INNER_PORT"`
}

type UserMgmtConfig struct {
	UserMgmtHost string `env:"USER_MGMT_HOST"`
	UserMgmtPort string `env:"USER_MGMT_PORT"`
}

type DbConfig struct {
	DatabaseName string `env:"DB_NAME"`
	UserName     string `env:"DB_USER"`
	Password     string `env:"DB_PASSWORD"`
	Host         string `env:"DB_HOST"`
	InnerPort    int    `env:"DB_INNER_PORT"`
	SslMode      string `env:"DB_SSL_MODE"`
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
