package repository

import (
	"fmt"

	"github.com/CHainGate/backend/internal/utils"
)

func InitAllRepositories() error {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", utils.Opts.DbHost, utils.Opts.DbUser, utils.Opts.DbPassword, utils.Opts.DbName, utils.Opts.DbPort)

	userRepo, err := NewUserRepository(dsn)
	if err != nil {
		return err
	}
	UserRepo = userRepo

	paymentRepo, err := NewPaymentRepository(dsn)
	if err != nil {
		return err
	}
	PaymentRepo = paymentRepo

	apiKeyRepo, err := NewApiKeyRepository(dsn)
	if err != nil {
		return err
	}
	ApiKeyRepo = apiKeyRepo
	return nil
}
