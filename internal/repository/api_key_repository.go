package repository

import (
	"github.com/CHainGate/backend/internal/model"
	"github.com/CHainGate/backend/pkg/enum"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type apiKeyRepository struct {
	DB *gorm.DB
}

type IApiKeyRepository interface {
	FindById(id string) (*model.ApiKey, error)
	FindByMerchantAndModeAndKeyType(merchantId uuid.UUID, mode enum.Mode, apiKeyType enum.ApiKeyType) (*model.ApiKey, error)
	Delete(merchantId uuid.UUID, apiKeyId string) error
}

func NewApiKeyRepository(db *gorm.DB) (IApiKeyRepository, error) {
	return &apiKeyRepository{db}, nil
}

func (r *apiKeyRepository) FindById(id string) (*model.ApiKey, error) {
	var apiKey model.ApiKey
	result := r.DB.Where("id = ?", id).Find(&apiKey)
	if result.Error != nil {
		return nil, result.Error
	}
	return &apiKey, nil
}

func (r *apiKeyRepository) FindByMerchantAndModeAndKeyType(merchantId uuid.UUID, mode enum.Mode, apiKeyType enum.ApiKeyType) (*model.ApiKey, error) {
	var key model.ApiKey
	result := r.DB.Where("merchant_id = ? and mode = ? and key_type = ?", merchantId, mode, apiKeyType).Find(&key)
	if result.Error != nil {
		return nil, result.Error
	}
	return &key, nil
}

func (r *apiKeyRepository) Delete(merchantId uuid.UUID, apiKeyId string) error {
	result := r.DB.Model(&model.ApiKey{}).Where("id = ? AND merchant_id = ?", apiKeyId, merchantId).Delete(&model.ApiKey{})
	if result.Error != nil {
		return result.Error
	}
	return nil
}
