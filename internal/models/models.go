package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	Id                uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4()"`
	FirstName         string
	LastName          string
	Email             string `gorm:"unique"`
	Password          []byte
	IsActive          bool
	CreatedAt         time.Time
	EmailVerification EmailVerification
	Wallets           []Wallet
	ApiKeys           []ApiKey
	Payments          []Payment
}

type EmailVerification struct {
	Id               uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4()"`
	UserId           uuid.UUID
	VerificationCode uint64
	CreatedAt        time.Time
}

type Wallet struct {
	Id        uuid.UUID `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	PaymentId uuid.UUID
	UserId    uuid.UUID
	Currency  string
	Mode      string
	address   string
}

type ApiKey struct {
	Id        uuid.UUID `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	UserId    uuid.UUID `gorm:"index:api_key_index,unique"`
	Mode      string    `gorm:"index:api_key_index,unique"`
	Key       string
	KeyType   string `gorm:"index:api_key_index,unique"`
	HashedKey string
	IsActive  bool `gorm:"index:api_key_index,unique,where:is_active = true"`
	CreatedAt time.Time
}

type Payment struct {
	BlockchainPaymentId uuid.UUID `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	UserId              uuid.UUID
	Wallet              Wallet
	Mode                string
	PriceAmount         float64 `gorm:"type:numeric"`
	PriceCurrency       string
	PayCurrency         string
	PayAddress          string
	CallbackUrl         string
	CreatedAt           time.Time
	UpdatedAt           time.Time
	PaymentStatus       []PaymentStatus `gorm:"foreignkey:PaymentId;references:BlockchainPaymentId"`
}

type PaymentStatus struct {
	Id            uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4()"`
	PaymentId     uuid.UUID
	PayAmount     float64 `gorm:"type:numeric"`
	ActuallyPaid  float64 `gorm:"type:numeric"`
	PaymentStatus string
	CreatedAt     time.Time
}