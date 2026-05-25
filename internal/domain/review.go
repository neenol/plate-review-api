package domain

import (
	"time"

	"github.com/google/uuid"
)

type Review struct {
	ID        uuid.UUID
	PlateID   uuid.UUID
	UserID    uuid.UUID
	Body      string
	Score     int64
	CreatedAt time.Time
	Author    *User
}
