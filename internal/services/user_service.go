package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/nglambertjr/plate-review-api/internal/domain"
)

type UserService struct {
	users UserRepository
}

func NewUserService(users UserRepository) *UserService {
	return &UserService{users: users}
}

func (s *UserService) Create(ctx context.Context, clerkUserID, username string) (*domain.User, error) {
	if username == "" {
		return nil, fmt.Errorf("username required: %w", domain.ErrInvalidInput)
	}
	u, err := s.users.Create(ctx, clerkUserID, username)
	if err != nil {
		if errors.Is(err, domain.ErrConflict) {
			return nil, fmt.Errorf("username or clerk ID already taken: %w", domain.ErrConflict)
		}
		return nil, fmt.Errorf("user service Create: %w", err)
	}
	return u, nil
}

func (s *UserService) GetByClerkID(ctx context.Context, clerkUserID string) (*domain.User, error) {
	u, err := s.users.GetByClerkID(ctx, clerkUserID)
	if err != nil {
		return nil, fmt.Errorf("user service GetByClerkID: %w", err)
	}
	return u, nil
}

func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	u, err := s.users.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user service GetByID: %w", err)
	}
	return u, nil
}

func (s *UserService) UpdateUsername(ctx context.Context, clerkUserID, newUsername string) (*domain.User, error) {
	if newUsername == "" {
		return nil, fmt.Errorf("username required: %w", domain.ErrInvalidInput)
	}
	u, err := s.users.GetByClerkID(ctx, clerkUserID)
	if err != nil {
		return nil, fmt.Errorf("user service UpdateUsername lookup: %w", err)
	}
	updated, err := s.users.UpdateUsername(ctx, u.ID, newUsername)
	if err != nil {
		if errors.Is(err, domain.ErrConflict) {
			return nil, fmt.Errorf("username already taken: %w", domain.ErrConflict)
		}
		return nil, fmt.Errorf("user service UpdateUsername: %w", err)
	}
	return updated, nil
}
