package database

import (
	"fmt"
	"github.com/CHainGate/backend/internal/models"
	"github.com/CHainGate/backend/internal/utils"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	DB *gorm.DB
)

func Connect() {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", utils.Opts.DbHost, utils.Opts.DbUser, utils.Opts.DbPassword, utils.Opts.DbName, utils.Opts.DbPort)
	connection, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("could not connect to the database")
	}

	err = connection.AutoMigrate(&models.User{})
	err = connection.AutoMigrate(&models.EmailVerification{})
	err = connection.AutoMigrate(&models.Wallet{})
	err = connection.AutoMigrate(&models.ApiKey{})
	err = connection.AutoMigrate(&models.Payment{})
	err = connection.AutoMigrate(&models.PaymentStatus{})

	DB = connection
}
