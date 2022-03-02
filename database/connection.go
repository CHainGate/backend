package database

import (
	"CHainGate/backend/model"
	"CHainGate/backend/utils"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", utils.Opts.DbHost, utils.Opts.DbUser, utils.Opts.DbPassword, utils.Opts.DbName, utils.Opts.DbPort)
	connection, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("could not connect to the database")
	}

	DB = connection

	err = connection.AutoMigrate(&model.User{})
	err = connection.AutoMigrate(&model.EmailVerification{})
	if err != nil {
		return
	}
}
