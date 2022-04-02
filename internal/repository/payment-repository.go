package repository

import (
	"github.com/CHainGate/backend/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	PaymentRepo IPaymentRepository
)

type IPaymentRepository interface {
	UpdatePayment(payment *models.Payment) error
	FindByBlockchainIdAndCurrency(id string, currency string) (*models.Payment, error)
}

type PaymentRepository struct {
	DB *gorm.DB
}

func NewPaymentRepository(dsn string) (IPaymentRepository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&models.Payment{})
	err = db.AutoMigrate(&models.PaymentStatus{})

	return &PaymentRepository{db}, nil
}

func (r *PaymentRepository) UpdatePayment(payment *models.Payment) error {
	result := r.DB.Save(&payment)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *PaymentRepository) FindByBlockchainIdAndCurrency(id string, currency string) (*models.Payment, error) {
	var payment models.Payment
	result := r.DB.Preload("PaymentStatus", func(db *gorm.DB) *gorm.DB {
		return db.Order("payment_statuses.created_at DESC")
	}).
		Where("blockchain_payment_id = ? AND pay_currency = ?", id, currency).
		First(&payment)
	if result.Error != nil {
		return nil, result.Error
	}
	return &payment, nil
}
