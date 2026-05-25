package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/nglambertjr/plate-review-api/internal/db/generated"
	"github.com/nglambertjr/plate-review-api/internal/domain"
)

type UserRepository struct {
	q *db.Queries
}

func NewUserRepository(dbtx db.DBTX) *UserRepository {
	return &UserRepository{q: db.New(dbtx)}
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	row, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("user repository GetByID: %w", err)
	}
	return mapUser(row), nil
}

func (r *UserRepository) GetByClerkID(ctx context.Context, clerkUserID string) (*domain.User, error) {
	row, err := r.q.GetUserByClerkID(ctx, clerkUserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("user repository GetByClerkID: %w", err)
	}
	return mapUser(row), nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	row, err := r.q.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("user repository GetByUsername: %w", err)
	}
	return mapUser(row), nil
}

func (r *UserRepository) Create(ctx context.Context, clerkUserID, username string) (*domain.User, error) {
	row, err := r.q.CreateUser(ctx, db.CreateUserParams{
		ClerkUserID: clerkUserID,
		Username:    username,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return nil, domain.ErrConflict
		}
		return nil, fmt.Errorf("user repository Create: %w", err)
	}
	return mapUser(row), nil
}

func (r *UserRepository) UpdateUsername(ctx context.Context, id uuid.UUID, username string) (*domain.User, error) {
	row, err := r.q.UpdateUserUsername(ctx, db.UpdateUserUsernameParams{
		ID:       id,
		Username: username,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		if isUniqueViolation(err) {
			return nil, domain.ErrConflict
		}
		return nil, fmt.Errorf("user repository UpdateUsername: %w", err)
	}
	return mapUser(row), nil
}

func mapUser(row db.User) *domain.User {
	return &domain.User{
		ID:          row.ID,
		ClerkUserID: row.ClerkUserID,
		Username:    row.Username,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
