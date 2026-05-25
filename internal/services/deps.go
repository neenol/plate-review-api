// deps.go declares all repository and platform interfaces consumed by the
// services in this package.  Multiple services can depend on the same interface
// (e.g. UserRepository) without duplicate declarations.  Each service file
// references the types declared here.
package services

import (
	"context"

	"github.com/google/uuid"

	"github.com/nglambertjr/plate-review-api/internal/domain"
)

// UserRepository is the persistence interface for user records.
type UserRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByClerkID(ctx context.Context, clerkUserID string) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	Create(ctx context.Context, clerkUserID, username string) (*domain.User, error)
	UpdateUsername(ctx context.Context, id uuid.UUID, username string) (*domain.User, error)
}

// PlateRepository is the persistence interface for plate records.
type PlateRepository interface {
	GetByHash(ctx context.Context, hash string) (*domain.Plate, error)
	Create(ctx context.Context, hash, state string) (*domain.Plate, error)
}

// ReviewRepository is the persistence interface for review records.
type ReviewRepository interface {
	Create(ctx context.Context, plateID, userID uuid.UUID, body string) (*domain.Review, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Review, error)
	ListByPlate(ctx context.Context, plateID uuid.UUID) ([]domain.Review, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetOwnerID(ctx context.Context, id uuid.UUID) (uuid.UUID, error)
}

// ReplyRepository is the persistence interface for reply records.
type ReplyRepository interface {
	Create(ctx context.Context, reviewID uuid.UUID, parentReplyID *uuid.UUID, userID uuid.UUID, body string) (*domain.Reply, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Reply, error)
	ListByReview(ctx context.Context, reviewID uuid.UUID) ([]domain.Reply, error)
	ListByParent(ctx context.Context, parentID uuid.UUID) ([]domain.Reply, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetOwnerID(ctx context.Context, id uuid.UUID) (uuid.UUID, error)
}

// VoteRepository is the persistence interface for vote records.
type VoteRepository interface {
	Create(ctx context.Context, userID uuid.UUID, targetType string, targetID uuid.UUID, value int32) (*domain.Vote, error)
	GetByUserAndTarget(ctx context.Context, userID uuid.UUID, targetType string, targetID uuid.UUID) (*domain.Vote, error)
	Delete(ctx context.Context, userID uuid.UUID, targetType string, targetID uuid.UUID) error
}

// AuthClient is the interface for JWT verification (implemented by platform/clerk).
type AuthClient interface {
	VerifyToken(token string) (clerkUserID string, err error)
}
