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
	Password          string
	Salt              []byte
	IsActive          bool
	EmailVerification EmailVerification
	Wallets           []Wallet
	ApiKeys           []ApiKey
	Payments          []Payment
}

type EmailVerification struct {
	Base
	MerchantId       uuid.UUID `gorm:"type:uuid"`
	VerificationCode uint64
}

type Wallet struct {
	Base
	MerchantId uuid.UUID           `gorm:"index:wallet_index,unique;type:uuid"`
	Currency   enum.CryptoCurrency `gorm:"index:wallet_index,unique"`
	Mode       enum.Mode           `gorm:"index:wallet_index,unique,where:deleted_at IS NULL"`
	Address    string
}

type ApiKey struct {
	Base
	MerchantId uuid.UUID       `gorm:"index:api_key_index,unique;type:uuid"`
	Mode       enum.Mode       `gorm:"index:api_key_index,unique"`
	KeyType    enum.ApiKeyType `gorm:"index:api_key_index,unique,where:deleted_at IS NULL""`
	ApiKey     string
	SecretKey  string
	Salt       []byte
}

type Payment struct {
	Base
	BlockchainPaymentId uuid.UUID `gorm:"type:uuid"`
	MerchantId          uuid.UUID `gorm:"type:uuid"`
	Wallet              Wallet
	WalletId            uuid.UUID `gorm:"type:uuid"`
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
	PaymentId    uuid.UUID `gorm:"type:uuid"`
	PayAmount    string
	ActuallyPaid string
	PaymentState enum.State
}
