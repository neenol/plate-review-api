package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID          uuid.UUID
	ClerkUserID string
	Username    string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
