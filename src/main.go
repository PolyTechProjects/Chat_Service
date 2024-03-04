package main

import (
	"example.com/main/src/database"
	"example.com/main/src/models"
	_ "github.com/lib/pq"
)

func main() {
	database.Init()
	db := database.DB
	db.AutoMigrate(&models.User{})

	defer db.Close()
}
