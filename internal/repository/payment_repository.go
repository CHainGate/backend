package repository

import (
	"github.com/CHainGate/backend/internal/model"
	"github.com/CHainGate/backend/pkg/enum"
	"gorm.io/gorm"
)

type paymentRepository struct {
	DB *gorm.DB
}

type IPaymentRepository interface {
	FindByBlockchainIdAndCurrency(id string, currency enum.CryptoCurrency) (*model.Payment, error)
	Update(payment *model.Payment) error
}

func NewPaymentRepository(db *gorm.DB) (IPaymentRepository, error) {
	return &paymentRepository{db}, nil
}

func (r *paymentRepository) FindByBlockchainIdAndCurrency(id string, currency enum.CryptoCurrency) (*model.Payment, error) {
	var payment model.Payment
	result := r.DB.Preload("PaymentStates", func(db *gorm.DB) *gorm.DB {
		return db.Order("payment_states.created_at DESC")
	}).
		Where("blockchain_payment_id = ? AND pay_currency = ?", id, currency).
		First(&payment)
	if result.Error != nil {
		return nil, result.Error
	}
	return &payment, nil
}

func (r *paymentRepository) Update(payment *model.Payment) error {
	result := r.DB.Save(&payment)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
