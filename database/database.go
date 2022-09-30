package database

import (
	"jwt-auth/models"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type databases interface {
	Connect(connString string)
	Migrate()
}

type database struct {
	Instance *gorm.DB
	dbError  error
}

var MainDB = &database{}
var SessionDB = &database{}

func (db *database) Connect(connString string, options *gorm.Config) {
	db.Instance, db.dbError = gorm.Open(mysql.Open(connString), options)
	if db.dbError != nil {
		log.Fatal(db.dbError)
		panic("Cannot connect to the database")
	}
	log.Println("Connected to the database")
}

func (db *database) Migrate() {
	err := db.Instance.AutoMigrate(&models.User{})
	if err != nil {
		log.Fatal(err)
		panic("Could not migrate defined models")
	}
}
