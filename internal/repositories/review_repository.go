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

type ReviewRepository struct {
	q *db.Queries
}

func NewReviewRepository(dbtx db.DBTX) *ReviewRepository {
	return &ReviewRepository{q: db.New(dbtx)}
}

func (r *ReviewRepository) Create(ctx context.Context, plateID, userID uuid.UUID, body string) (*domain.Review, error) {
	row, err := r.q.CreateReview(ctx, db.CreateReviewParams{
		PlateID: plateID,
		UserID:  userID,
		Body:    body,
	})
	if err != nil {
		return nil, fmt.Errorf("review repository Create: %w", err)
	}
	return &domain.Review{
		ID:        row.ID,
		PlateID:   row.PlateID,
		UserID:    row.UserID,
		Body:      row.Body,
		CreatedAt: row.CreatedAt.Time,
	}, nil
}

func (r *ReviewRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Review, error) {
	row, err := r.q.GetReviewByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("review repository GetByID: %w", err)
	}
	return mapReviewRow(row), nil
}

func (r *ReviewRepository) ListByPlate(ctx context.Context, plateID uuid.UUID) ([]domain.Review, error) {
	rows, err := r.q.ListReviewsByPlate(ctx, plateID)
	if err != nil {
		return nil, fmt.Errorf("review repository ListByPlate: %w", err)
	}
	out := make([]domain.Review, len(rows))
	for i, row := range rows {
		out[i] = *mapListReviewRow(row)
	}
	return out, nil
}

func (r *ReviewRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.q.DeleteReview(ctx, id); err != nil {
		return fmt.Errorf("review repository Delete: %w", err)
	}
	return nil
}

func (r *ReviewRepository) GetOwnerID(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	ownerID, err := r.q.GetReviewOwnerID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.UUID{}, domain.ErrNotFound
		}
		return uuid.UUID{}, fmt.Errorf("review repository GetOwnerID: %w", err)
	}
	return ownerID, nil
}

func mapReviewRow(row db.GetReviewByIDRow) *domain.Review {
	return &domain.Review{
		ID:        row.ID,
		PlateID:   row.PlateID,
		UserID:    row.UserID,
		Body:      row.Body,
		Score:     row.Score,
		CreatedAt: row.CreatedAt.Time,
		Author:    &domain.User{ID: row.UserID, Username: row.AuthorUsername},
	}
}

func mapListReviewRow(row db.ListReviewsByPlateRow) *domain.Review {
	return &domain.Review{
		ID:        row.ID,
		PlateID:   row.PlateID,
		UserID:    row.UserID,
		Body:      row.Body,
		Score:     row.Score,
		CreatedAt: row.CreatedAt.Time,
		Author:    &domain.User{ID: row.UserID, Username: row.AuthorUsername},
	}
}
