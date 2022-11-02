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
	numOfRetries := 6
	timeoutSec := 15
	for {
		if numOfRetries <= 0 {
			panic(fmt.Sprintf("Could not connect to the database after %ds", timeoutSec*numOfRetries))
		}

		db.Instance, db.dbError = gorm.Open(mysql.Open(connString), options)
		fmt.Println("Trying to connect: ", connString)
		if db.dbError != nil {
			time.Sleep(time.Duration(timeoutSec) * time.Second)
			numOfRetries--
			continue
		}

		log.Println("Connected to the database")
		break
	}
}

func SaveUser(user *models.User) error {
	result := MainDB.Instance.Create(&user)
	return result.Error
}
