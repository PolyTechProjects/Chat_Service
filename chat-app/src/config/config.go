package config

import (
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env      string         `yaml:"env"`
	Web      WebConfig      `yaml:"web"`
	Database DatabaseConfig `yaml:"db"`
}

type WebConfig struct {
	Port int `yaml:"port"`
}

type DatabaseConfig struct {
	Port    int    `yaml:"port"`
	Host    string `yaml:"host"`
	Sslmode string `yaml:"sslmode"`
}

func MustLoad() *Config {
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		panic("config path is empty")
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		dir, err := os.Getwd()
		panic("config file does not exist: " + path + " pwd: " + dir + err.Error())
	}
	var cfg Config
	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		panic("failed to read config: " + err.Error())
	}
	return &cfg
}
