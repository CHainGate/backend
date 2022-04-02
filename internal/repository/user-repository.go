package repository

import (
	"github.com/CHainGate/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	UserRepo IUserRepository
)

type IUserRepository interface {
	FindByEmail(email string) (*models.User, error)
	FindById(id uuid.UUID) (*models.User, error)
	CreateUser(user *models.User) error
	UpdateUser(user *models.User) error
	CreateWallet(wallet *models.Wallet) error
}

type UserRepository struct {
	DB *gorm.DB
}

func NewUserRepository(dsn string) (IUserRepository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&models.User{})
	err = db.AutoMigrate(&models.EmailVerification{})
	err = db.AutoMigrate(&models.Wallet{})

	return &UserRepository{db}, nil
}

func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	result := r.DB.Where("email = ?", email).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (r *UserRepository) FindById(id uuid.UUID) (*models.User, error) {
	var user models.User
	result := r.DB.Preload("Wallets").Where("id = ?", id).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (r *UserRepository) CreateUser(user *models.User) error {
	result := r.DB.Create(&user)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *UserRepository) UpdateUser(user *models.User) error {
	result := r.DB.Save(&user)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *UserRepository) CreateWallet(wallet *models.Wallet) error {
	result := r.DB.Create(&wallet)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
