package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/nglambertjr/plate-review-api/internal/domain"
)

type ReplyService struct {
	replies ReplyRepository
	reviews ReviewRepository
	users   UserRepository
}

func NewReplyService(replies ReplyRepository, reviews ReviewRepository, users UserRepository) *ReplyService {
	return &ReplyService{replies: replies, reviews: reviews, users: users}
}

// CreateOnReview creates a top-level reply on a review.
func (s *ReplyService) CreateOnReview(ctx context.Context, clerkUserID string, reviewID uuid.UUID, body string) (*domain.Reply, error) {
	body = strings.TrimSpace(body)
	if body == "" {
		return nil, fmt.Errorf("body required: %w", domain.ErrInvalidInput)
	}

	user, err := s.users.GetByClerkID(ctx, clerkUserID)
	if err != nil {
		return nil, fmt.Errorf("reply service CreateOnReview user: %w", err)
	}

	if _, err = s.reviews.GetOwnerID(ctx, reviewID); err != nil {
		return nil, fmt.Errorf("reply service CreateOnReview review: %w", err)
	}

	reply, err := s.replies.Create(ctx, reviewID, nil, user.ID, body)
	if err != nil {
		return nil, fmt.Errorf("reply service CreateOnReview: %w", err)
	}
	reply.Author = &domain.User{ID: user.ID, Username: user.Username}
	return reply, nil
}

// CreateOnReply creates a nested reply under an existing reply.
func (s *ReplyService) CreateOnReply(ctx context.Context, clerkUserID string, parentReplyID uuid.UUID, body string) (*domain.Reply, error) {
	body = strings.TrimSpace(body)
	if body == "" {
		return nil, fmt.Errorf("body required: %w", domain.ErrInvalidInput)
	}

	user, err := s.users.GetByClerkID(ctx, clerkUserID)
	if err != nil {
		return nil, fmt.Errorf("reply service CreateOnReply user: %w", err)
	}

	parent, err := s.replies.GetByID(ctx, parentReplyID)
	if err != nil {
		return nil, fmt.Errorf("reply service CreateOnReply parent: %w", err)
	}

	reply, err := s.replies.Create(ctx, parent.ReviewID, &parentReplyID, user.ID, body)
	if err != nil {
		return nil, fmt.Errorf("reply service CreateOnReply: %w", err)
	}
	reply.Author = &domain.User{ID: user.ID, Username: user.Username}
	return reply, nil
}

func (s *ReplyService) ListByReview(ctx context.Context, reviewID uuid.UUID) ([]domain.Reply, error) {
	replies, err := s.replies.ListByReview(ctx, reviewID)
	if err != nil {
		return nil, fmt.Errorf("reply service ListByReview: %w", err)
	}
	return replies, nil
}

func (s *ReplyService) ListByParent(ctx context.Context, parentReplyID uuid.UUID) ([]domain.Reply, error) {
	replies, err := s.replies.ListByParent(ctx, parentReplyID)
	if err != nil {
		return nil, fmt.Errorf("reply service ListByParent: %w", err)
	}
	return replies, nil
}

func (s *ReplyService) Delete(ctx context.Context, clerkUserID string, replyID uuid.UUID) error {
	user, err := s.users.GetByClerkID(ctx, clerkUserID)
	if err != nil {
		return fmt.Errorf("reply service Delete user: %w", err)
	}

	ownerID, err := s.replies.GetOwnerID(ctx, replyID)
	if err != nil {
		return fmt.Errorf("reply service Delete owner: %w", err)
	}

	if ownerID != user.ID {
		return domain.ErrForbidden
	}

	return s.replies.Delete(ctx, replyID)
}
