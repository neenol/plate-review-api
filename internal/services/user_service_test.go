package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/nglambertjr/plate-review-api/internal/domain"
	"github.com/nglambertjr/plate-review-api/internal/services"
)

func TestUserService_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("creates user successfully", func(t *testing.T) {
		repo := newFakeUserRepo()
		svc := services.NewUserService(repo)

		u, err := svc.Create(ctx, "clerk_123", "alice")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if u.Username != "alice" {
			t.Errorf("username: want alice, got %s", u.Username)
		}
		if u.ClerkUserID != "clerk_123" {
			t.Errorf("clerk_user_id: want clerk_123, got %s", u.ClerkUserID)
		}
	})

	t.Run("empty username returns invalid input", func(t *testing.T) {
		repo := newFakeUserRepo()
		svc := services.NewUserService(repo)

		_, err := svc.Create(ctx, "clerk_123", "")
		if !errors.Is(err, domain.ErrInvalidInput) {
			t.Errorf("want ErrInvalidInput, got %v", err)
		}
	})

	t.Run("duplicate username returns conflict", func(t *testing.T) {
		repo := newFakeUserRepo()
		svc := services.NewUserService(repo)

		if _, err := svc.Create(ctx, "clerk_1", "alice"); err != nil {
			t.Fatalf("first create: %v", err)
		}
		_, err := svc.Create(ctx, "clerk_2", "alice")
		if !errors.Is(err, domain.ErrConflict) {
			t.Errorf("want ErrConflict, got %v", err)
		}
	})
}

func TestUserService_GetByClerkID(t *testing.T) {
	ctx := context.Background()
	repo := newFakeUserRepo()
	svc := services.NewUserService(repo)

	if _, err := svc.Create(ctx, "clerk_abc", "bob"); err != nil {
		t.Fatalf("setup: %v", err)
	}

	u, err := svc.GetByClerkID(ctx, "clerk_abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.Username != "bob" {
		t.Errorf("want bob, got %s", u.Username)
	}

	_, err = svc.GetByClerkID(ctx, "clerk_missing")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}

func TestUserService_UpdateUsername(t *testing.T) {
	ctx := context.Background()

	t.Run("successful rename", func(t *testing.T) {
		repo := newFakeUserRepo()
		svc := services.NewUserService(repo)
		if _, err := svc.Create(ctx, "clerk_x", "original"); err != nil {
			t.Fatalf("setup: %v", err)
		}

		u, err := svc.UpdateUsername(ctx, "clerk_x", "renamed")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if u.Username != "renamed" {
			t.Errorf("want renamed, got %s", u.Username)
		}
	})

	t.Run("username taken returns conflict", func(t *testing.T) {
		repo := newFakeUserRepo()
		svc := services.NewUserService(repo)
		if _, err := svc.Create(ctx, "clerk_a", "taken"); err != nil {
			t.Fatalf("setup: %v", err)
		}
		if _, err := svc.Create(ctx, "clerk_b", "renamer"); err != nil {
			t.Fatalf("setup: %v", err)
		}

		_, err := svc.UpdateUsername(ctx, "clerk_b", "taken")
		if !errors.Is(err, domain.ErrConflict) {
			t.Errorf("want ErrConflict, got %v", err)
		}
	})
}
