package repository

import (
	"fmt"

	"github.com/CHainGate/backend/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/CHainGate/backend/internal/utils"
)

func SetupDatabase() (IMerchantRepository, IApiKeyRepository, IPaymentRepository, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", utils.Opts.DbHost, utils.Opts.DbUser, utils.Opts.DbPassword, utils.Opts.DbName, utils.Opts.DbPort)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, nil, nil, err
	}

	err = autoMigrateDB(db)
	if err != nil {
		return nil, nil, nil, err
	}

	merchantRepo, apiKeyRepo, paymentRepo, err := createRepositories(db)
	if err != nil {
		return nil, nil, nil, err
	}

	return merchantRepo, apiKeyRepo, paymentRepo, nil
}

func autoMigrateDB(db *gorm.DB) error {
	err := db.AutoMigrate(&model.Merchant{})
	if err != nil {
		return err
	}
	err = db.AutoMigrate(&model.EmailVerification{})
	if err != nil {
		return err
	}
	err = db.AutoMigrate(&model.Wallet{})
	if err != nil {
		return err
	}
	err = db.AutoMigrate(&model.ApiKey{})
	if err != nil {
		return err
	}
	err = db.AutoMigrate(&model.Payment{})
	if err != nil {
		return err
	}
	err = db.AutoMigrate(&model.PaymentState{})
	if err != nil {
		return err
	}
	return nil
}

func createRepositories(db *gorm.DB) (IMerchantRepository, IApiKeyRepository, IPaymentRepository, error) {
	merchantRepo, err := NewMerchantRepository(db)
	if err != nil {
		return nil, nil, nil, err
	}

	paymentRepo, err := NewPaymentRepository(db)
	if err != nil {
		return nil, nil, nil, err
	}

	apiKeyRepo, err := NewApiKeyRepository(db)
	if err != nil {
		return nil, nil, nil, err
	}
	return merchantRepo, apiKeyRepo, paymentRepo, nil
}
