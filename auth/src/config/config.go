package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	App      AppConfig
	UserMgmt UserMgmtConfig
	Db       DbConfig
	Keycloak KeycloakConfig
	Redis    RedisConfig
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
	InnerPort    int    `env:"DB_PORT"`
	SslMode      string `env:"DB_SSL_MODE"`
}

type KeycloakConfig struct {
	Host         string `env:"KEYCLOAK_HOST"`
	InnerPort    int    `env:"KEYCLOAK_PORT"`
	Realm        string `env:"KEYCLOAK_REALM"`
	ClientId     string `env:"KEYCLOAK_CLIENT_ID"`
	ClientSecret string `env:"KEYCLOAK_CLIENT_SECRET"`
}

type RedisConfig struct {
	Db          int    `env:"REDIS_DB"`
	Password    string `env:"REDIS_PASSWORD"`
	Host        string `env:"REDIS_HOST"`
	InnerPort   int    `env:"REDIS_PORT"`
	ChannelName string `env:"REDIS_CHANNEL_NAME"`
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
