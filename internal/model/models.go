package model

import (
	"database/sql/driver"
	"fmt"
	"math/big"
	"reflect"
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
	PayAmount    *BigInt   `gorm:"type:numeric"`
	ActuallyPaid *BigInt   `gorm:"type:numeric"`
	PaymentState enum.State
}

type BigInt struct {
	big.Int
}

func NewBigIntFromInt(value int64) *BigInt {
	x := new(big.Int).SetInt64(value)
	return NewBigInt(x)
}

func NewBigIntFromString(value string) *BigInt {
	x, ok := new(big.Int).SetString(value, 10)
	if !ok {
		fmt.Println("SetString: error")
		return NewBigIntFromInt(0)
	}
	return NewBigInt(x)
}

func NewBigInt(value *big.Int) *BigInt {
	return &BigInt{Int: *value}
}

func (bigInt *BigInt) Value() (driver.Value, error) {
	if bigInt == nil {
		return "null", nil
	}
	return bigInt.String(), nil
}

func (bigInt *BigInt) Scan(val interface{}) error {
	if val == nil {
		return nil
	}
	var data string
	switch v := val.(type) {
	case []byte:
		data = string(v)
	case string:
		data = v
	case int64:
		*bigInt = *NewBigIntFromInt(v)
		return nil
	default:
		return fmt.Errorf("bigint: can't convert %s type to *big.Int", reflect.TypeOf(val).Kind())
	}
	bigI, ok := new(big.Int).SetString(data, 10)
	if !ok {
		return fmt.Errorf("not a valid big integer: %s", data)
	}
	bigInt.Int = *bigI
	return nil
}
