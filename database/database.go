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

type database[T models.User | models.Session] struct {
	Instance *gorm.DB
	model    T
	dbError  error
}

var MainDB = &database[models.User]{}
var SessionDB = &database[models.Session]{}

func (db *database[T]) Connect(connString string, options *gorm.Config) {
	db.Instance, db.dbError = gorm.Open(mysql.Open(connString), options)
	if db.dbError != nil {
		log.Fatal(db.dbError)
		panic("Cannot connect to the database")
	}
	log.Println("Connected to the database")
}

func (db *database[T]) Migrate(model *T) {
	err := db.Instance.AutoMigrate(model)
	if err != nil {
		log.Fatal(err)
		panic("Could not migrate defined models")
	}
}
