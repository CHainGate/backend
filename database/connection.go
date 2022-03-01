package database

import (
	"CHainGate/backend/model"
	"CHainGate/backend/utils"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"time"
)

var DB *gorm.DB

func Connect() {
	retries := 10
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", utils.Opts.DbHost, utils.Opts.DbUser, utils.Opts.DbPassword, utils.Opts.DbName, utils.Opts.DbPort)
	connection, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	for err != nil {
		if retries > 1 {
			retries--
			time.Sleep(5 * time.Second)
			connection, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
			continue
		}
		panic("could not connect to the database")
	}

	DB = connection

	err = connection.AutoMigrate(&model.User{})
	if err != nil {
		return
	}
}
