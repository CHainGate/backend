package repository

import (
	"github.com/CHainGate/backend/internal/models"
	"github.com/CHainGate/backend/internal/utils"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	ApiKeyRepo IApiKeyRepository
)

type IApiKeyRepository interface {
	FindApiKeyByUserModeKeyType(userId uuid.UUID, mode utils.Mode, apiKeyType utils.ApiKeyType) ([]models.ApiKey, error)
	DeleteApiKey(userId uuid.UUID, apiKeyId string) error
	FindApiKeyById(id string) (*models.ApiKey, error)
}

type ApiKeyRepository struct {
	DB *gorm.DB
}

func NewApiKeyRepository(dsn string) (IApiKeyRepository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&models.ApiKey{})

	return &ApiKeyRepository{db}, nil
}

func (r *ApiKeyRepository) FindApiKeyByUserModeKeyType(userId uuid.UUID, mode utils.Mode, apiKeyType utils.ApiKeyType) ([]models.ApiKey, error) {
	var keys []models.ApiKey
	result := r.DB.Where("user_id = ? and mode = ? and key_type = ?", userId, mode.String(), apiKeyType.String()).Find(&keys)
	if result.Error != nil {
		return nil, result.Error
	}
	return keys, nil
}

func (r *ApiKeyRepository) DeleteApiKey(userId uuid.UUID, apiKeyId string) error {
	result := r.DB.Model(&models.ApiKey{}).Where("id = ? AND user_id = ?", apiKeyId, userId).Update("is_active", false)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *ApiKeyRepository) FindApiKeyById(id string) (*models.ApiKey, error) {
	var apiKey models.ApiKey
	result := r.DB.Where("id = ?", id).Find(&apiKey)
	if result.Error != nil {
		return nil, result.Error
	}
	return &apiKey, nil
}
