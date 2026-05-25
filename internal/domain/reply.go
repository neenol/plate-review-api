package domain

import (
	"time"

	"github.com/google/uuid"
)

type Reply struct {
	ID            uuid.UUID
	ReviewID      uuid.UUID
	ParentReplyID *uuid.UUID
	UserID        uuid.UUID
	Body          string
	Score         int64
	CreatedAt     time.Time
	Author        *User
}
