package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/nglambertjr/plate-review-api/internal/domain"
	"github.com/nglambertjr/plate-review-api/internal/services"
)

const testPepper = "test-pepper"

func TestReviewService_Create(t *testing.T) {
	ctx := context.Background()

	setup := func() (*services.ReviewService, *fakeUserRepo) {
		userRepo := newFakeUserRepo()
		plateRepo := newFakePlateRepo()
		reviewRepo := newFakeReviewRepo()
		svc := services.NewReviewService(reviewRepo, userRepo, plateRepo, testPepper)
		return svc, userRepo
	}

	t.Run("creates review successfully", func(t *testing.T) {
		svc, userRepo := setup()
		userRepo.seed(&domain.User{ID: uuid.New(), ClerkUserID: "clerk_1", Username: "alice"})

		rev, err := svc.Create(ctx, "clerk_1", "CA", "abc123", "great driver")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rev.Body != "great driver" {
			t.Errorf("body mismatch: %s", rev.Body)
		}
		if rev.Author == nil || rev.Author.Username != "alice" {
			t.Error("author not populated")
		}
	})

	t.Run("empty body returns invalid input", func(t *testing.T) {
		svc, userRepo := setup()
		userRepo.seed(&domain.User{ID: uuid.New(), ClerkUserID: "clerk_1", Username: "alice"})

		_, err := svc.Create(ctx, "clerk_1", "CA", "abc123", "")
		if !errors.Is(err, domain.ErrInvalidInput) {
			t.Errorf("want ErrInvalidInput, got %v", err)
		}
	})

	t.Run("unknown user returns not found", func(t *testing.T) {
		svc, _ := setup()

		_, err := svc.Create(ctx, "clerk_missing", "CA", "abc123", "body")
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("want ErrNotFound, got %v", err)
		}
	})

	t.Run("invalid plate returns invalid input", func(t *testing.T) {
		svc, userRepo := setup()
		userRepo.seed(&domain.User{ID: uuid.New(), ClerkUserID: "clerk_1", Username: "alice"})

		_, err := svc.Create(ctx, "clerk_1", "XX", "abc123", "body")
		if !errors.Is(err, domain.ErrInvalidInput) {
			t.Errorf("want ErrInvalidInput, got %v", err)
		}
	})
}

func TestReviewService_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("owner can delete", func(t *testing.T) {
		userRepo := newFakeUserRepo()
		plateRepo := newFakePlateRepo()
		reviewRepo := newFakeReviewRepo()
		svc := services.NewReviewService(reviewRepo, userRepo, plateRepo, testPepper)

		u := &domain.User{ID: uuid.New(), ClerkUserID: "clerk_1", Username: "alice"}
		userRepo.seed(u)

		rev, err := svc.Create(ctx, "clerk_1", "CA", "abc123", "body")
		if err != nil {
			t.Fatalf("create: %v", err)
		}

		if err = svc.Delete(ctx, "clerk_1", rev.ID); err != nil {
			t.Fatalf("delete: %v", err)
		}
	})

	t.Run("non-owner gets forbidden", func(t *testing.T) {
		userRepo := newFakeUserRepo()
		plateRepo := newFakePlateRepo()
		reviewRepo := newFakeReviewRepo()
		svc := services.NewReviewService(reviewRepo, userRepo, plateRepo, testPepper)

		owner := &domain.User{ID: uuid.New(), ClerkUserID: "clerk_owner", Username: "owner"}
		other := &domain.User{ID: uuid.New(), ClerkUserID: "clerk_other", Username: "other"}
		userRepo.seed(owner)
		userRepo.seed(other)

		rev, err := svc.Create(ctx, "clerk_owner", "CA", "abc123", "body")
		if err != nil {
			t.Fatalf("create: %v", err)
		}

		err = svc.Delete(ctx, "clerk_other", rev.ID)
		if !errors.Is(err, domain.ErrForbidden) {
			t.Errorf("want ErrForbidden, got %v", err)
		}
	})
}

func TestReviewService_List(t *testing.T) {
	ctx := context.Background()
	userRepo := newFakeUserRepo()
	plateRepo := newFakePlateRepo()
	reviewRepo := newFakeReviewRepo()
	svc := services.NewReviewService(reviewRepo, userRepo, plateRepo, testPepper)

	userRepo.seed(&domain.User{ID: uuid.New(), ClerkUserID: "clerk_1", Username: "alice"})

	// No plate exists yet — should return empty slice, not error.
	reviews, err := svc.List(ctx, "CA", "ABC123")
	if err != nil {
		t.Fatalf("list before plate exists: %v", err)
	}
	if len(reviews) != 0 {
		t.Errorf("want 0 reviews, got %d", len(reviews))
	}

	if _, err = svc.Create(ctx, "clerk_1", "CA", "ABC123", "first"); err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err = svc.Create(ctx, "clerk_1", "CA", "ABC123", "second"); err != nil {
		t.Fatalf("create: %v", err)
	}

	reviews, err = svc.List(ctx, "CA", "ABC123")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(reviews) != 2 {
		t.Errorf("want 2 reviews, got %d", len(reviews))
	}
}
