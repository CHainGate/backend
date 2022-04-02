package userRepository

import (
	"github.com/CHainGate/backend/internal/models"
	"github.com/CHainGate/backend/internal/utils"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	Repository IUserRepository
)

type IUserRepository interface {
	FindByEmail(email string) (*models.User, error)
	FindById(id uuid.UUID) (*models.User, error)
	CreateUser(user *models.User) error
	UpdateUser(user *models.User) error
	FindApiKeyByUserModeKeyType(userId uuid.UUID, mode utils.Mode, apiKeyType utils.ApiKeyType) ([]models.ApiKey, error)
	DeleteApiKey(userId uuid.UUID, apiKeyId string) error
	FindApiKeyById(id string) (*models.ApiKey, error)
	CreateWallet(wallet *models.Wallet) error
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

func (r *UserRepository) FindApiKeyByUserModeKeyType(userId uuid.UUID, mode utils.Mode, apiKeyType utils.ApiKeyType) ([]models.ApiKey, error) {
	var keys []models.ApiKey
	result := r.DB.Where("user_id = ? and mode = ? and key_type = ?", userId, mode.String(), apiKeyType.String()).Find(&keys)
	if result.Error != nil {
		return nil, result.Error
	}
	return keys, nil
}

func (r *UserRepository) DeleteApiKey(userId uuid.UUID, apiKeyId string) error {
	result := r.DB.Model(&models.ApiKey{}).Where("id = ? AND user_id = ?", apiKeyId, userId).Update("is_active", false)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *UserRepository) FindApiKeyById(id string) (*models.ApiKey, error) {
	var apiKey models.ApiKey
	result := r.DB.Where("id = ?", id).Find(&apiKey)
	if result.Error != nil {
		return nil, result.Error
	}
	return &apiKey, nil
}

func (r *UserRepository) CreateWallet(wallet *models.Wallet) error {
	result := r.DB.Create(&wallet)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
