package model

import (
	"time"

	"github.com/CHainGate/backend/pkg/enum"
	"gorm.io/gorm"

	"github.com/google/uuid"
)

type Base struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type Merchant struct {
	Base
	FirstName         string
	LastName          string
	Email             string `gorm:"unique"`
	Password          []byte
	IsActive          bool
	EmailVerification EmailVerification
	Wallets           []Wallet
	ApiKeys           []ApiKey
	Payments          []Payment
}

type EmailVerification struct {
	Base
	MerchantId       uuid.UUID
	VerificationCode uint64
}

type Wallet struct {
	Base
	MerchantId uuid.UUID
	Currency   enum.CryptoCurrency
	Mode       enum.Mode
	Address    string
}

type ApiKey struct {
	Base
	MerchantId uuid.UUID       `gorm:"index:api_key_index,unique"`
	Mode       enum.Mode       `gorm:"index:api_key_index,unique"`
	KeyType    enum.ApiKeyType `gorm:"index:api_key_index,unique"`
	ApiKey     string
	SecretKey  string
	Salt       []byte
	IsActive   bool `gorm:"index:api_key_index,unique,where:is_active = true"`
}

type Payment struct {
	Base
	BlockchainPaymentId uuid.UUID
	MerchantId          uuid.UUID
	Wallet              Wallet
	WalletId            uuid.UUID
	Mode                enum.Mode
	PriceAmount         float64 `gorm:"type:numeric"`
	PriceCurrency       enum.FiatCurrency
	PayCurrency         enum.CryptoCurrency
	PayAddress          string
	CallbackUrl         string
	PaymentStates       []PaymentState
}

type PaymentState struct {
	Base
	PaymentId    uuid.UUID
	PayAmount    float64 `gorm:"type:numeric"`
	ActuallyPaid float64 `gorm:"type:numeric"`
	PaymentState enum.State
}
