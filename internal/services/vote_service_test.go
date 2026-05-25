package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/nglambertjr/plate-review-api/internal/domain"
	"github.com/nglambertjr/plate-review-api/internal/services"
)

func TestVoteService_CastAndRetract(t *testing.T) {
	ctx := context.Background()

	makeServices := func() (*services.VoteService, *fakeUserRepo, *fakeReviewRepo) {
		userRepo := newFakeUserRepo()
		reviewRepo := newFakeReviewRepo()
		replyRepo := newFakeReplyRepo()
		voteRepo := newFakeVoteRepo()
		svc := services.NewVoteService(voteRepo, userRepo, reviewRepo, replyRepo)
		return svc, userRepo, reviewRepo
	}

	t.Run("cast vote on review", func(t *testing.T) {
		svc, userRepo, reviewRepo := makeServices()

		u := &domain.User{ID: uuid.New(), ClerkUserID: "clerk_1", Username: "alice"}
		userRepo.seed(u)
		rev := &domain.Review{ID: uuid.New(), UserID: u.ID}
		reviewRepo.reviews[rev.ID] = rev

		vote, err := svc.CastVote(ctx, "clerk_1", domain.VoteTargetReview, rev.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if vote.Value != 1 {
			t.Errorf("want value=1, got %d", vote.Value)
		}
	})

	t.Run("duplicate vote returns conflict", func(t *testing.T) {
		svc, userRepo, reviewRepo := makeServices()

		u := &domain.User{ID: uuid.New(), ClerkUserID: "clerk_2", Username: "bob"}
		userRepo.seed(u)
		rev := &domain.Review{ID: uuid.New(), UserID: u.ID}
		reviewRepo.reviews[rev.ID] = rev

		if _, err := svc.CastVote(ctx, "clerk_2", domain.VoteTargetReview, rev.ID); err != nil {
			t.Fatalf("first vote: %v", err)
		}
		_, err := svc.CastVote(ctx, "clerk_2", domain.VoteTargetReview, rev.ID)
		if !errors.Is(err, domain.ErrConflict) {
			t.Errorf("want ErrConflict, got %v", err)
		}
	})

	t.Run("retract vote", func(t *testing.T) {
		svc, userRepo, reviewRepo := makeServices()

		u := &domain.User{ID: uuid.New(), ClerkUserID: "clerk_3", Username: "carol"}
		userRepo.seed(u)
		rev := &domain.Review{ID: uuid.New(), UserID: u.ID}
		reviewRepo.reviews[rev.ID] = rev

		if _, err := svc.CastVote(ctx, "clerk_3", domain.VoteTargetReview, rev.ID); err != nil {
			t.Fatalf("cast: %v", err)
		}
		if err := svc.RetractVote(ctx, "clerk_3", domain.VoteTargetReview, rev.ID); err != nil {
			t.Fatalf("retract: %v", err)
		}
	})

	t.Run("retract non-existent vote returns not found", func(t *testing.T) {
		svc, userRepo, _ := makeServices()
		u := &domain.User{ID: uuid.New(), ClerkUserID: "clerk_4", Username: "dave"}
		userRepo.seed(u)

		err := svc.RetractVote(ctx, "clerk_4", domain.VoteTargetReview, uuid.New())
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("want ErrNotFound, got %v", err)
		}
	})

	t.Run("invalid target type returns invalid input", func(t *testing.T) {
		svc, userRepo, _ := makeServices()
		u := &domain.User{ID: uuid.New(), ClerkUserID: "clerk_5", Username: "eve"}
		userRepo.seed(u)

		_, err := svc.CastVote(ctx, "clerk_5", "bad_type", uuid.New())
		if !errors.Is(err, domain.ErrInvalidInput) {
			t.Errorf("want ErrInvalidInput, got %v", err)
		}
	})
}
