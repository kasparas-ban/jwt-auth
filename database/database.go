package database

import (
	"fmt"
	"jwt-auth/models"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type databases interface {
	Connect(connString string)
	Migrate()
}

type database[T models.User | Session] struct {
	Instance *gorm.DB
	model    T
	dbError  error
}

var MainDB = &database[models.User]{}
var SessionDB = &database[Session]{}

func (db *database[T]) Migrate(model *T) {
	err := db.Instance.AutoMigrate(model)
	if err != nil {
		panic("Could not migrate defined models")
	}
}

func (db *database[T]) Connect(connString string, options *gorm.Config) {
	i := 6
	for {
		if i <= 0 {
			panic(fmt.Sprintf("Could not connect to the database after %ds", 15*6))
		}

		db.Instance, db.dbError = gorm.Open(mysql.Open(connString), options)
		fmt.Println("Trying to connect: ", connString)
		if db.dbError != nil {
			time.Sleep(15 * time.Second)
			i--
			continue
		}

		log.Println("Connected to the database")
		break
	}
}
