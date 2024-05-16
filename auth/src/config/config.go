package config

import (
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env  string     `yaml:"env"`
	GRPC GRPCConfig `yaml:"grpc"`
	DB   DbConfig   `yaml:"db"`
}

type GRPCConfig struct {
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

type DbConfig struct {
	Host    string `yaml:"host"`
	Port    string `yaml:"port"`
	SslMode string `yaml:"sslmode"`
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
