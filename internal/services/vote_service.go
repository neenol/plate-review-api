package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/nglambertjr/plate-review-api/internal/domain"
)

type VoteService struct {
	votes   VoteRepository
	users   UserRepository
	reviews ReviewRepository
	replies ReplyRepository
}

func NewVoteService(votes VoteRepository, users UserRepository, reviews ReviewRepository, replies ReplyRepository) *VoteService {
	return &VoteService{votes: votes, users: users, reviews: reviews, replies: replies}
}

// CastVote upvotes the target (review or reply) on behalf of the authenticated user.
// Returns ErrConflict if the user has already voted.
func (s *VoteService) CastVote(ctx context.Context, clerkUserID, targetType string, targetID uuid.UUID) (*domain.Vote, error) {
	if targetType != domain.VoteTargetReview && targetType != domain.VoteTargetReply {
		return nil, fmt.Errorf("invalid target type %q: %w", targetType, domain.ErrInvalidInput)
	}

	user, err := s.users.GetByClerkID(ctx, clerkUserID)
	if err != nil {
		return nil, fmt.Errorf("vote service CastVote user: %w", err)
	}

	if err = s.verifyTargetExists(ctx, targetType, targetID); err != nil {
		return nil, err
	}

	vote, err := s.votes.Create(ctx, user.ID, targetType, targetID, 1)
	if err != nil {
		if errors.Is(err, domain.ErrConflict) {
			return nil, fmt.Errorf("already voted: %w", domain.ErrConflict)
		}
		return nil, fmt.Errorf("vote service CastVote: %w", err)
	}
	return vote, nil
}

// RetractVote removes the user's vote from the target.
// Returns ErrNotFound if no vote exists.
func (s *VoteService) RetractVote(ctx context.Context, clerkUserID, targetType string, targetID uuid.UUID) error {
	if targetType != domain.VoteTargetReview && targetType != domain.VoteTargetReply {
		return fmt.Errorf("invalid target type %q: %w", targetType, domain.ErrInvalidInput)
	}

	user, err := s.users.GetByClerkID(ctx, clerkUserID)
	if err != nil {
		return fmt.Errorf("vote service RetractVote user: %w", err)
	}

	if _, err = s.votes.GetByUserAndTarget(ctx, user.ID, targetType, targetID); err != nil {
		return fmt.Errorf("vote service RetractVote lookup: %w", err)
	}

	return s.votes.Delete(ctx, user.ID, targetType, targetID)
}

func (s *VoteService) verifyTargetExists(ctx context.Context, targetType string, targetID uuid.UUID) error {
	var err error
	switch targetType {
	case domain.VoteTargetReview:
		_, err = s.reviews.GetOwnerID(ctx, targetID)
	case domain.VoteTargetReply:
		_, err = s.replies.GetOwnerID(ctx, targetID)
	}
	if err != nil {
		return fmt.Errorf("vote target: %w", err)
	}
	return nil
}
