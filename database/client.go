package database

import (
	"jwt-auth/models"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var Instance *gorm.DB
var dbError error

func Connect(connString string) {
	Instance, dbError = gorm.Open(mysql.Open(connString), &gorm.Config{})
	if dbError != nil {
		log.Fatal(dbError)
		panic("Cannot connect to the database")
	}
	log.Println("Connected to the database")
}

func Migrate() {
	err := Instance.AutoMigrate(&models.User{})
	if err != nil {
		log.Fatal(err)
		panic("Could not migrate defined models")
	}
}
