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
	MerchantId uuid.UUID           `gorm:"type:uuid"`
	Currency   enum.CryptoCurrency `gorm:"type:varchar"`
	Mode       enum.Mode           `gorm:"type:varchar"`
	Address    string
}

type ApiKey struct {
	Base
	MerchantId uuid.UUID       `gorm:"index:api_key_index,unique;type:uuid"`
	Mode       enum.Mode       `gorm:"index:api_key_index,unique;type:varchar"`
	KeyType    enum.ApiKeyType `gorm:"index:api_key_index,unique;type:varchar"`
	ApiKey     string
	SecretKey  string
	Salt       []byte
	IsActive   bool `gorm:"index:api_key_index,unique,where:is_active = true"`
}

type Payment struct {
	Base
	BlockchainPaymentId uuid.UUID `gorm:"type:uuid"`
	MerchantId          uuid.UUID `gorm:"type:uuid"`
	Wallet              Wallet
	WalletId            uuid.UUID           `gorm:"type:uuid"`
	Mode                enum.Mode           `gorm:"type:varchar"`
	PriceAmount         float64             `gorm:"type:numeric"`
	PriceCurrency       enum.FiatCurrency   `gorm:"type:varchar"`
	PayCurrency         enum.CryptoCurrency `gorm:"type:varchar"`
	PayAddress          string
	CallbackUrl         string
	PaymentStates       []PaymentState
}

type PaymentState struct {
	Base
	PaymentId    uuid.UUID  `gorm:"type:uuid"`
	PayAmount    float64    `gorm:"type:numeric"`
	ActuallyPaid float64    `gorm:"type:numeric"`
	PaymentState enum.State `gorm:"type:varchar"`
}
