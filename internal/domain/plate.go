package domain

import (
	"time"

	"github.com/google/uuid"
)

type Plate struct {
	ID        uuid.UUID
	PlateHash string
	State     string
	CreatedAt time.Time
}
