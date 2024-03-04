package database

import (
	"os"

	"example.com/m/models"
	"github.com/jinzhu/gorm"
)

var DB *gorm.DB

type DbConnectStruct struct {
	host     string
	port     string
	user     string
	password string
	dbname   string
	sslmode  string
}

func Init() {
	db_struct := DbConnectStruct{
		host:     os.Getenv("DB_HOST"),
		port:     os.Getenv("DB_PORT"),
		user:     os.Getenv("DB_USER"),
		password: os.Getenv("DB_PASSWORD"),
		dbname:   os.Getenv("DB_NAME"),
		sslmode:  os.Getenv("DB_SSLMODE"),
	}
	db, err := gorm.Open(
		"postgres",
		"host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		db_struct.host,
		db_struct.port,
		db_struct.user,
		db_struct.dbname,
		db_struct.password,
		db_struct.sslmode,
	)
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&models.User{})
}
