package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/nglambertjr/plate-review-api/internal/db/generated"
	"github.com/nglambertjr/plate-review-api/internal/domain"
)

type PlateRepository struct {
	q *db.Queries
}

func NewPlateRepository(dbtx db.DBTX) *PlateRepository {
	return &PlateRepository{q: db.New(dbtx)}
}

func (r *PlateRepository) GetByHash(ctx context.Context, hash string) (*domain.Plate, error) {
	row, err := r.q.GetPlateByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("plate repository GetByHash: %w", err)
	}
	return mapPlate(row), nil
}

func (r *PlateRepository) Create(ctx context.Context, hash, state string) (*domain.Plate, error) {
	row, err := r.q.CreatePlate(ctx, db.CreatePlateParams{
		PlateHash: hash,
		State:     state,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return nil, domain.ErrConflict
		}
		return nil, fmt.Errorf("plate repository Create: %w", err)
	}
	return mapPlate(row), nil
}

func mapPlate(row db.Plate) *domain.Plate {
	return &domain.Plate{
		ID:        row.ID,
		PlateHash: row.PlateHash,
		State:     row.State,
		CreatedAt: row.CreatedAt.Time,
	}
}
