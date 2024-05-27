package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	App       AppConfig
	Auth      AuthConfig
	UserMgmt  UserMgmtConfig
	Db        DbConfig
	SeaweedFS SeaweedFSConfig
	Redis     RedisConfig
}

type AppConfig struct {
	HttpInnerPort int `env:"APP_HTTP_INNER_PORT"`
	GrpcInnerPort int `env:"APP_GRPC_INNER_PORT"`
}

type AuthConfig struct {
	AuthHost string `env:"AUTH_HOST"`
	AuthPort string `env:"AUTH_PORT"`
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

type SeaweedFSConfig struct {
	MasterIp   string `env:"SEAWEEDFS_MASTER_HOST"`
	MasterPort int    `env:"SEAWEEDFS_MASTER_PORT"`
}

type RedisConfig struct {
	Db        int    `env:"REDIS_DB"`
	Password  string `env:"REDIS_PASSWORD"`
	Host      string `env:"REDIS_HOST"`
	InnerPort int    `env:"REDIS_INNER_PORT"`
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
