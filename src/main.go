package main

import (
	"example.com/m/models"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
)

func main() {
	// postgress
	db, err := gorm.Open("postgres", "host=localhost port=5432 user=postgres dbname=postgres password=postgres sslmode=disable")
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&models.User{})

	defer db.Close()
}
