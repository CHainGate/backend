package websocket

import (
	"github.com/CHainGate/backend/pkg/enum"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

// We'll need to define an Upgrader
// this will require a Read and Write buffer size
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type Client struct {
	ID   string
	Conn *websocket.Conn
	Pool *Pool
}

type Message struct {
	Type        string      `json:"type"`
	MessageType string      `json:"messageType"`
	Body        interface{} `json:"body"`
}

func (c *Client) SendInitialCoins() {
	message := Message{MessageType: "currencies", Body: enum.GetCryptoCurrencyDetails()}
	c.Conn.WriteJSON(message)
}

func (c *Client) SendWaiting() {
	message := Message{MessageType: "wait-for-tx", Body: enum.GetCryptoCurrencyDetails()}
	c.Pool.Broadcast <- message
}

func (c *Client) SendReceivedTX() {
	message := Message{MessageType: "received-tx", Body: enum.GetCryptoCurrencyDetails()}
	c.Pool.Broadcast <- message
}

func (c *Client) SendConfirmed() {
	message := Message{MessageType: "confirmed", Body: enum.GetCryptoCurrencyDetails()}
	c.Pool.Broadcast <- message
}

func (c *Client) Read() {
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
		selected := mapCurrency["currency"]
		log.Println("cs", selected)
		break
	}
}

func Upgrade(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return conn, nil
}
