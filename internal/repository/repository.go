package repository

import (
	"fmt"

	"github.com/CHainGate/backend/internal/repository/userRepository"
	"github.com/CHainGate/backend/internal/utils"
)

func InitAllRepositories() error {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", utils.Opts.DbHost, utils.Opts.DbUser, utils.Opts.DbPassword, utils.Opts.DbName, utils.Opts.DbPort)

	repository, err := userRepository.NewRepository(dsn)
	if err != nil {
		return err
	}
	userRepository.Repository = repository

	return nil
}
