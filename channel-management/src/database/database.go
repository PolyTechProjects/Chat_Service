package database

import (
	"fmt"
	"log/slog"
	"os"

	"example.com/channel-management/src/config"
	"example.com/channel-management/src/internal/models"
	"github.com/jinzhu/gorm"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var DB *gorm.DB

func Init(cfg *config.Config) {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
	slog.Info("Connecting to DB")
	str := fmt.Sprintf(
		"postgres://%v:%v@%v:%v/%v?sslmode=%v",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		cfg.Database.Host,
		cfg.Database.InnerPort,
		os.Getenv("DB_NAME"),
		cfg.Database.SslMode,
	)
	slog.Info(str)
	db, err := gorm.Open(
		"postgres",
		str,
	)
	if err != nil {
		slog.Error("Error has occured while connecting to DB", err)
		panic(err)
	}

	db.AutoMigrate(&models.Channel{}, &models.UserChannel{}, &models.Admin{})
	DB = db
	slog.Info("Connected to DB")
}

func Close() {
	slog.Info("Disconneting from DB")
	DB.Close()
}
