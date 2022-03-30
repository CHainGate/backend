package userRepository

import (
	"github.com/CHainGate/backend/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	Repository IUserRepository
)

type IUserRepository interface {
	FindByEmail(id string) (*models.User, error)
	CreateUser(user *models.User) error
	UpdateUser(user *models.User) error
}

type UserRepository struct {
	DB *gorm.DB
}

func NewRepository(dsn string) (IUserRepository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// if we need one repository per model move autoMigrate to specific repository
	err = db.AutoMigrate(&models.User{})
	err = db.AutoMigrate(&models.EmailVerification{})
	err = db.AutoMigrate(&models.Wallet{})
	err = db.AutoMigrate(&models.ApiKey{})
	err = db.AutoMigrate(&models.Payment{})
	err = db.AutoMigrate(&models.PaymentStatus{})

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
