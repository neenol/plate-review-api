package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	VoteTargetReview = "review"
	VoteTargetReply  = "reply"
)

type Vote struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	TargetType string
	TargetID   uuid.UUID
	Value      int32
	CreatedAt  time.Time
}
