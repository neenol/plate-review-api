package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/nglambertjr/plate-review-api/internal/db/generated"
	"github.com/nglambertjr/plate-review-api/internal/domain"
)

type VoteRepository struct {
	q *db.Queries
}

func NewVoteRepository(dbtx db.DBTX) *VoteRepository {
	return &VoteRepository{q: db.New(dbtx)}
}

func (r *VoteRepository) Create(ctx context.Context, userID uuid.UUID, targetType string, targetID uuid.UUID, value int32) (*domain.Vote, error) {
	row, err := r.q.CreateVote(ctx, db.CreateVoteParams{
		UserID:     userID,
		TargetType: targetType,
		TargetID:   targetID,
		Value:      value,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return nil, domain.ErrConflict
		}
		return nil, fmt.Errorf("vote repository Create: %w", err)
	}
	return mapVote(row), nil
}

func (r *VoteRepository) GetByUserAndTarget(ctx context.Context, userID uuid.UUID, targetType string, targetID uuid.UUID) (*domain.Vote, error) {
	row, err := r.q.GetVote(ctx, db.GetVoteParams{
		UserID:     userID,
		TargetType: targetType,
		TargetID:   targetID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("vote repository GetByUserAndTarget: %w", err)
	}
	return mapVote(row), nil
}

func (r *VoteRepository) Delete(ctx context.Context, userID uuid.UUID, targetType string, targetID uuid.UUID) error {
	if err := r.q.DeleteVote(ctx, db.DeleteVoteParams{
		UserID:     userID,
		TargetType: targetType,
		TargetID:   targetID,
	}); err != nil {
		return fmt.Errorf("vote repository Delete: %w", err)
	}
	return nil
}

func mapVote(row db.Vote) *domain.Vote {
	return &domain.Vote{
		ID:         row.ID,
		UserID:     row.UserID,
		TargetType: row.TargetType,
		TargetID:   row.TargetID,
		Value:      row.Value,
		CreatedAt:  row.CreatedAt.Time,
	}
}
