package config

import (
	"github.com/CHainGate/backend/internal/model"
	"github.com/google/uuid"
)

var Pools = make(map[uuid.UUID]*model.Pool)
