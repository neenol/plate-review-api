package controllers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/nglambertjr/plate-review-api/internal/controllers"
	"github.com/nglambertjr/plate-review-api/internal/domain"
)

type stubReviewService struct {
	createFn  func(ctx context.Context, clerkUserID, state, rawPlate, body string) (*domain.Review, error)
	getByIDFn func(ctx context.Context, id uuid.UUID) (*domain.Review, error)
	listFn    func(ctx context.Context, state, rawPlate string) ([]domain.Review, error)
	deleteFn  func(ctx context.Context, clerkUserID string, reviewID uuid.UUID) error
}

func (s *stubReviewService) Create(ctx context.Context, clerkUserID, state, rawPlate, body string) (*domain.Review, error) {
	if s.createFn != nil {
		return s.createFn(ctx, clerkUserID, state, rawPlate, body)
	}
	return nil, errors.New("not configured")
}

func (s *stubReviewService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Review, error) {
	if s.getByIDFn != nil {
		return s.getByIDFn(ctx, id)
	}
	return nil, errors.New("not configured")
}

func (s *stubReviewService) List(ctx context.Context, state, rawPlate string) ([]domain.Review, error) {
	if s.listFn != nil {
		return s.listFn(ctx, state, rawPlate)
	}
	return nil, errors.New("not configured")
}

func (s *stubReviewService) Delete(ctx context.Context, clerkUserID string, reviewID uuid.UUID) error {
	if s.deleteFn != nil {
		return s.deleteFn(ctx, clerkUserID, reviewID)
	}
	return errors.New("not configured")
}

func TestReviewController_Create(t *testing.T) {
	revID := uuid.New()
	svc := &stubReviewService{
		createFn: func(_ context.Context, clerkUserID, state, _, body string) (*domain.Review, error) {
			return &domain.Review{
				ID:        revID,
				Body:      body,
				CreatedAt: time.Now(),
				Author:    &domain.User{Username: "alice"},
			}, nil
		},
	}

	ctrl := controllers.NewReviewController(svc)
	r := chi.NewRouter()
	r.With(injectAuth("clerk_1")).Post("/plates/{state}/{plate}/reviews", ctrl.Create)

	body := bytes.NewBufferString(`{"body":"good driver"}`)
	req := httptest.NewRequest(http.MethodPost, "/plates/CA/ABC123/reviews", body)
	req.Header.Set("Authorization", "Bearer token")
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("want 201, got %d; body: %s", rr.Code, rr.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp["body"] != "good driver" {
		t.Errorf("body mismatch: %v", resp["body"])
	}
}

func TestReviewController_List(t *testing.T) {
	svc := &stubReviewService{
		listFn: func(_ context.Context, _, _ string) ([]domain.Review, error) {
			return []domain.Review{
				{ID: uuid.New(), Body: "first", CreatedAt: time.Now(), Author: &domain.User{Username: "a"}},
				{ID: uuid.New(), Body: "second", CreatedAt: time.Now(), Author: &domain.User{Username: "b"}},
			}, nil
		},
	}

	ctrl := controllers.NewReviewController(svc)
	r := chi.NewRouter()
	r.Get("/plates/{state}/{plate}/reviews", ctrl.List)

	req := httptest.NewRequest(http.MethodGet, "/plates/TX/XYZ999/reviews", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("want 200, got %d", rr.Code)
	}
	var resp []map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(resp) != 2 {
		t.Errorf("want 2 reviews, got %d", len(resp))
	}
}

func TestReviewController_Delete(t *testing.T) {
	revID := uuid.New()

	t.Run("success returns 204", func(t *testing.T) {
		svc := &stubReviewService{
			deleteFn: func(_ context.Context, _ string, _ uuid.UUID) error { return nil },
		}
		ctrl := controllers.NewReviewController(svc)
		r := chi.NewRouter()
		r.With(injectAuth("clerk_1")).Delete("/reviews/{id}", ctrl.Delete)

		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/reviews/%s", revID), nil)
		req.Header.Set("Authorization", "Bearer token")
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusNoContent {
			t.Errorf("want 204, got %d", rr.Code)
		}
	})

	t.Run("forbidden returns 403", func(t *testing.T) {
		svc := &stubReviewService{
			deleteFn: func(_ context.Context, _ string, _ uuid.UUID) error { return domain.ErrForbidden },
		}
		ctrl := controllers.NewReviewController(svc)
		r := chi.NewRouter()
		r.With(injectAuth("clerk_1")).Delete("/reviews/{id}", ctrl.Delete)

		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/reviews/%s", revID), nil)
		req.Header.Set("Authorization", "Bearer token")
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusForbidden {
			t.Errorf("want 403, got %d", rr.Code)
		}
	})
}
