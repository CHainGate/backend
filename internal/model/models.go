package model

import (
	"database/sql/driver"
	"fmt"
	"log"
	"math/big"
	"reflect"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/exp/slices"

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
	MerchantId uuid.UUID `gorm:"index:api_key_index,unique;type:uuid"`
	Mode       enum.Mode `gorm:"index:api_key_index,unique"`
	ApiKey     string
	Secret     string
	SecretSalt []byte
}

type Payment struct {
	Base
	BlockchainPaymentId uuid.UUID `gorm:"type:uuid"`
	MerchantId          uuid.UUID `gorm:"type:uuid"`
	Wallet              *Wallet
	WalletId            *uuid.UUID `gorm:"type:uuid"`
	Mode                enum.Mode
	PriceAmount         float64 `gorm:"type:numeric"`
	PriceCurrency       enum.FiatCurrency
	PayCurrency         enum.CryptoCurrency
	PayAddress          string
	CallbackUrl         string
	SuccessPageUrl      string
	FailurePageUrl      string
	TxHash              string
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

type Message struct {
	Type        string      `json:"type"`
	MessageType string      `json:"messageType"`
	Body        interface{} `json:"body"`
}
type Pool struct {
	Register   chan *Client
	Unregister chan *Client
	Clients    map[*Client]bool
	Broadcast  chan Message
}

type Client struct {
	ID   string
	Conn *websocket.Conn
	Pool *Pool
}

type SocketBody struct {
	Currency       string    `json:"currency"`
	PayAddress     string    `json:"payAddress"`
	PayAmount      string    `json:"payAmount"`
	ExpireTime     time.Time `json:"expireTime"`
	Mode           string    `json:"mode"`
	SuccessPageURL string    `json:"successPageURL"`
	FailurePageURL string    `json:"failurePageURL"`
}

func (c *Client) SendInitialCoins() {
	message := Message{MessageType: enum.CurrencySelection.String(), Body: enum.GetCryptoCurrencyDetails()}
	c.Conn.WriteJSON(message)
}

func GetWaitingCreateDate(payment *Payment) time.Time {
	index := slices.IndexFunc(payment.PaymentStates, func(ps PaymentState) bool { return ps.PaymentState == enum.Waiting })
	return payment.PaymentStates[index].CreatedAt
}

func (c *Client) SendWaiting(p *Payment) {
	body := SocketBody{
		Currency:       p.PayCurrency.String(),
		PayAddress:     p.PayAddress,
		PayAmount:      p.PaymentStates[0].PayAmount.String(),
		ExpireTime:     GetWaitingCreateDate(p).Add(15 * time.Minute),
		Mode:           p.Mode.String(),
		SuccessPageURL: p.SuccessPageUrl,
		FailurePageURL: p.FailurePageUrl,
	}
	message := Message{MessageType: enum.Waiting.String(), Body: body}
	c.Pool.Broadcast <- message
}

func (c *Client) SendReceivedTX() {
	message := Message{MessageType: enum.Paid.String(), Body: enum.GetCryptoCurrencyDetails()}
	c.Pool.Broadcast <- message
}

func (c *Client) SendConfirmed() {
	message := Message{MessageType: enum.Confirmed.String(), Body: enum.GetCryptoCurrencyDetails()}
	c.Pool.Broadcast <- message
}

func (c *Client) SendExpired() {
	message := Message{MessageType: enum.Expired.String(), Body: enum.GetCryptoCurrencyDetails()}
	c.Pool.Broadcast <- message
}

func (c *Client) SendFailed() {
	message := Message{MessageType: enum.Failed.String(), Body: enum.GetCryptoCurrencyDetails()}
	c.Pool.Broadcast <- message
}

func (c *Client) Read() string {
	selected := ""
	for {
		var message Message
		err := c.Conn.ReadJSON(&message)
		if err != nil {
			log.Println("read failed:", err)
			c.Pool.Unregister <- c
			c.Conn.Close()
			break
		}
		mapCurrency := message.Body.(map[string]interface{})
		selected = mapCurrency["currency"].(string)
		break
	}
	return selected
}

func (pool *Pool) Start() {
	for {
		select {
		case client := <-pool.Register:
			pool.Clients[client] = true
			fmt.Println("Size of Connection Pool: ", len(pool.Clients))
			break
		case client := <-pool.Unregister:
			delete(pool.Clients, client)
			fmt.Println("Size of Connection Pool: ", len(pool.Clients))
			break
		case message := <-pool.Broadcast:
			fmt.Println("Sending message to all clients in Pool")
			for client, _ := range pool.Clients {
				if err := client.Conn.WriteJSON(message); err != nil {
					fmt.Println(err)
					client.Pool.Unregister <- client
					client.Conn.Close()
				}
			}
		}
	}
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
