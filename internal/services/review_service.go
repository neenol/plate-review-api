package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/nglambertjr/plate-review-api/internal/domain"
	"github.com/nglambertjr/plate-review-api/internal/plate"
)

type ReviewService struct {
	reviews ReviewRepository
	users   UserRepository
	plates  PlateRepository
	pepper  string
}

func NewReviewService(reviews ReviewRepository, users UserRepository, plates PlateRepository, pepper string) *ReviewService {
	return &ReviewService{reviews: reviews, users: users, plates: plates, pepper: pepper}
}

func (s *ReviewService) Create(ctx context.Context, clerkUserID, state, rawPlate, body string) (*domain.Review, error) {
	body = strings.TrimSpace(body)
	if body == "" {
		return nil, fmt.Errorf("body required: %w", domain.ErrInvalidInput)
	}

	user, err := s.users.GetByClerkID(ctx, clerkUserID)
	if err != nil {
		return nil, fmt.Errorf("review service Create user lookup: %w", err)
	}

	_, hash, err := plate.NormalizeAndHash(state, rawPlate, s.pepper)
	if err != nil {
		return nil, err
	}

	p, err := s.getOrCreatePlate(ctx, hash, state)
	if err != nil {
		return nil, fmt.Errorf("review service Create plate: %w", err)
	}

	rev, err := s.reviews.Create(ctx, p.ID, user.ID, body)
	if err != nil {
		return nil, fmt.Errorf("review service Create: %w", err)
	}
	rev.Author = &domain.User{ID: user.ID, Username: user.Username}
	return rev, nil
}

func (s *ReviewService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Review, error) {
	rev, err := s.reviews.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("review service GetByID: %w", err)
	}
	return rev, nil
}

func (s *ReviewService) List(ctx context.Context, state, rawPlate string) ([]domain.Review, error) {
	_, hash, err := plate.NormalizeAndHash(state, rawPlate, s.pepper)
	if err != nil {
		return nil, err
	}

	p, err := s.plates.GetByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return []domain.Review{}, nil
		}
		return nil, fmt.Errorf("review service List plate: %w", err)
	}

	reviews, err := s.reviews.ListByPlate(ctx, p.ID)
	if err != nil {
		return nil, fmt.Errorf("review service List: %w", err)
	}
	return reviews, nil
}

func (s *ReviewService) Delete(ctx context.Context, clerkUserID string, reviewID uuid.UUID) error {
	user, err := s.users.GetByClerkID(ctx, clerkUserID)
	if err != nil {
		return fmt.Errorf("review service Delete user lookup: %w", err)
	}

	ownerID, err := s.reviews.GetOwnerID(ctx, reviewID)
	if err != nil {
		return fmt.Errorf("review service Delete owner lookup: %w", err)
	}

	if ownerID != user.ID {
		return domain.ErrForbidden
	}

	return s.reviews.Delete(ctx, reviewID)
}

func (s *ReviewService) getOrCreatePlate(ctx context.Context, hash, state string) (*domain.Plate, error) {
	p, err := s.plates.GetByHash(ctx, hash)
	if err == nil {
		return p, nil
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}
	p, err = s.plates.Create(ctx, hash, strings.ToUpper(state))
	if err != nil {
		if errors.Is(err, domain.ErrConflict) {
			return s.plates.GetByHash(ctx, hash)
		}
		return nil, err
	}
	return p, nil
}
