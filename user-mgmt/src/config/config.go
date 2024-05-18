package config

import (
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	App  AppConfig
	Auth AuthConfig
	Db   DbConfig
}

type AppConfig struct {
	InnerPort int `env:"APP_INNER_PORT"`
}

type AuthConfig struct {
	AuthHost string `env:"AUTH_HOST"`
	AuthPort string `env:"AUTH_PORT"`
}

type DbConfig struct {
	DatabaseName string `env:"DB_NAME"`
	UserName     string `env:"DB_USER"`
	Password     string `env:"DB_PASSWORD"`
	Host         string `env:"DB_HOST"`
	InnerPort    int    `env:"DB_INNER_PORT"`
}

func MustLoad() *Config {
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		panic("failed to read config: " + err.Error())
	}
	return &cfg
}
