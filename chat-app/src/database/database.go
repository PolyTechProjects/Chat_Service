package database

import (
	"fmt"
	"log/slog"
	"os"

	"example.com/chat-app/src/config"
	"example.com/chat-app/src/internal/models"
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
		"host=%v port=%v user=%v dbname=%v password=%v sslmode=%v",
		cfg.Database.Host,
		cfg.Database.Port,
		os.Getenv("DB_USER"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PASSWORD"),
		cfg.Database.Sslmode,
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

	db.AutoMigrate(&models.ChatRoomXUser{}, &models.Message{})
	DB = db
	slog.Info("Connected to DB")
}

func Close() {
	slog.Info("Disconneting from DB")
	DB.Close()
}
