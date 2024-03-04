package main

import (
	"log"

	"example.com/main/src/database"
	"example.com/main/src/models"
	_ "github.com/lib/pq"
)

func main() {
	database.Init()
	db := database.DB
	db.AutoMigrate(&models.User{})
	defer log.Default().Printf("DB :%v", db.HasTable(&models.User{}))
	defer log.Default().Printf("Program successfully finished!")
	defer db.Close()
}
