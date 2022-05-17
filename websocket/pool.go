package websocket

import (
	"github.com/CHainGate/backend/internal/model"
)

func NewPool() *model.Pool {
	return &model.Pool{
		Register:   make(chan *model.Client),
		Unregister: make(chan *model.Client),
		Clients:    make(map[*model.Client]bool),
		Broadcast:  make(chan model.Message),
	}
}
