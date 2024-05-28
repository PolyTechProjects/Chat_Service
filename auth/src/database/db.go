package database

import (
	"fmt"
	"log"
	"log/slog"

	"example.com/main/src/config"
	"example.com/main/src/models"
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
	str := fmt.Sprintf(
		"postgres://%v:%v@%v:%v/%v?sslmode=%v",
		cfg.Db.UserName,
		cfg.Db.Password,
		cfg.Db.Host,
		cfg.Db.InnerPort,
		cfg.Db.DatabaseName,
		cfg.Db.SslMode,
	)
	db, err := gorm.Open(
		"postgres",
		str,
	)
	if err != nil {
		log.Panicln(err, str)
		panic(err.Error())
	}
	db.AutoMigrate(&models.User{})
	DB = db
	slog.Debug("Connected to DB")
}
