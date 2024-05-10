package main

import (
	"log"
	"log/slog"
	"time"

	"example.com/main/src/database"
	"example.com/main/src/internal/app"
	"example.com/main/src/models"
	_ "github.com/lib/pq"
)

func main() {
	database.Init()
	db := database.DB
	db.AutoMigrate(&models.User{})
	defer log.Default().Printf("Program successfully finished!")
	defer db.Close()

	application := app.New(
		slog.Default(),
		8080,
		"db",
		24*time.Hour,
	)
	application.GRPCSrv.Run()
}
