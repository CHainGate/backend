package repository

import (
	"github.com/CHainGate/backend/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	MerchantRepository IMerchantRepository
)

type merchantRepository struct {
	DB *gorm.DB
}

type IMerchantRepository interface {
	FindById(id uuid.UUID) (*model.Merchant, error)
	FindByEmail(email string) (*model.Merchant, error)
	Create(merchant *model.Merchant) error
	Update(merchant *model.Merchant) error
	CreateWallet(wallet *model.Wallet) error
}

func NewMerchantRepository(db *gorm.DB) (IMerchantRepository, error) {
	return &merchantRepository{db}, nil
}

func (r *merchantRepository) FindById(id uuid.UUID) (*model.Merchant, error) {
	var merchant model.Merchant
	result := r.DB.
		Preload("EmailVerification").
		Preload("Wallets").
		Where("id = ?", id).
		First(&merchant)
	if result.Error != nil {
		return nil, result.Error
	}
	return &merchant, nil
}

func (r *merchantRepository) FindByEmail(email string) (*model.Merchant, error) {
	var merchant model.Merchant
	result := r.DB.Preload("EmailVerification").Where("email = ?", email).First(&merchant)
	if result.Error != nil {
		return nil, result.Error
	}
	return &merchant, nil
}

func (r *merchantRepository) Create(merchant *model.Merchant) error {
	result := r.DB.Create(&merchant)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *merchantRepository) Update(merchant *model.Merchant) error {
	result := r.DB.Save(&merchant)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *merchantRepository) CreateWallet(wallet *model.Wallet) error {
	result := r.DB.Create(&wallet)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
