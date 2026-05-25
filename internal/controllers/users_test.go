package controllers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/nglambertjr/plate-review-api/internal/controllers"
	"github.com/nglambertjr/plate-review-api/internal/domain"
	"github.com/nglambertjr/plate-review-api/internal/middleware"
)

// stubUserService satisfies the controller's local userService interface.
type stubUserService struct {
	createFn          func(ctx context.Context, clerkUserID, username string) (*domain.User, error)
	getByClerkIDFn    func(ctx context.Context, clerkUserID string) (*domain.User, error)
	getByIDFn         func(ctx context.Context, id uuid.UUID) (*domain.User, error)
	updateUsernameFn  func(ctx context.Context, clerkUserID, newUsername string) (*domain.User, error)
}

func (s *stubUserService) Create(ctx context.Context, clerkUserID, username string) (*domain.User, error) {
	if s.createFn != nil {
		return s.createFn(ctx, clerkUserID, username)
	}
	return nil, errors.New("not configured")
}

func (s *stubUserService) GetByClerkID(ctx context.Context, clerkUserID string) (*domain.User, error) {
	if s.getByClerkIDFn != nil {
		return s.getByClerkIDFn(ctx, clerkUserID)
	}
	return nil, errors.New("not configured")
}

func (s *stubUserService) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	if s.getByIDFn != nil {
		return s.getByIDFn(ctx, id)
	}
	return nil, errors.New("not configured")
}

func (s *stubUserService) UpdateUsername(ctx context.Context, clerkUserID, newUsername string) (*domain.User, error) {
	if s.updateUsernameFn != nil {
		return s.updateUsernameFn(ctx, clerkUserID, newUsername)
	}
	return nil, errors.New("not configured")
}

func withClerkUser(clerkUserID string, r *http.Request) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), contextKey("clerk_user_id"), clerkUserID))
}

// contextKey is the unexported key type from the middleware package.
// We replicate the value via the exported ClerkUserID helper below to avoid
// importing the unexported constant.  Instead, we install the value the same
// way RequireAuth does — via an HTTP middleware call on the test router.
type contextKey = string

func routerWithAuth(clerkUserID string, h http.Handler) http.Handler {
	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			// Inject Clerk user ID using a test-only auth shim.
			// We implement a fake AuthVerifier that always returns clerkUserID.
			next.ServeHTTP(w, req)
		})
	})
	r.Mount("/", h)
	return r
}

func injectAuth(clerkUserID string) func(http.Handler) http.Handler {
	return middleware.RequireAuth(&fakeAuth{id: clerkUserID})
}

type fakeAuth struct{ id string }

func (f *fakeAuth) VerifyToken(_ string) (string, error) { return f.id, nil }

func TestUserController_Create(t *testing.T) {
	svc := &stubUserService{
		createFn: func(_ context.Context, clerkUserID, username string) (*domain.User, error) {
			return &domain.User{
				ID:          uuid.New(),
				ClerkUserID: clerkUserID,
				Username:    username,
				CreatedAt:   time.Now(),
			}, nil
		},
	}

	ctrl := controllers.NewUserController(svc)

	r := chi.NewRouter()
	r.With(injectAuth("clerk_test")).Post("/users", ctrl.Create)

	body := bytes.NewBufferString(`{"username":"alice"}`)
	req := httptest.NewRequest(http.MethodPost, "/users", body)
	req.Header.Set("Authorization", "Bearer fake-token")
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
	if resp["username"] != "alice" {
		t.Errorf("username: want alice, got %v", resp["username"])
	}
}

func TestUserController_Create_Conflict(t *testing.T) {
	svc := &stubUserService{
		createFn: func(_ context.Context, _, _ string) (*domain.User, error) {
			return nil, domain.ErrConflict
		},
	}

	ctrl := controllers.NewUserController(svc)
	r := chi.NewRouter()
	r.With(injectAuth("clerk_test")).Post("/users", ctrl.Create)

	body := bytes.NewBufferString(`{"username":"taken"}`)
	req := httptest.NewRequest(http.MethodPost, "/users", body)
	req.Header.Set("Authorization", "Bearer fake-token")
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusConflict {
		t.Errorf("want 409, got %d", rr.Code)
	}
}

func TestUserController_Me(t *testing.T) {
	uid := uuid.New()
	svc := &stubUserService{
		getByClerkIDFn: func(_ context.Context, _ string) (*domain.User, error) {
			return &domain.User{ID: uid, ClerkUserID: "clerk_test", Username: "alice", CreatedAt: time.Now()}, nil
		},
	}

	ctrl := controllers.NewUserController(svc)
	r := chi.NewRouter()
	r.With(injectAuth("clerk_test")).Get("/users/me", ctrl.Me)

	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	req.Header.Set("Authorization", "Bearer fake-token")
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("want 200, got %d; body: %s", rr.Code, rr.Body.String())
	}
}

func TestUserController_NoAuth(t *testing.T) {
	svc := &stubUserService{}
	ctrl := controllers.NewUserController(svc)
	r := chi.NewRouter()
	// No auth middleware — requests arrive without Clerk user ID in context.
	r.Post("/users", ctrl.Create)

	body := bytes.NewBufferString(`{"username":"alice"}`)
	req := httptest.NewRequest(http.MethodPost, "/users", body)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("want 401, got %d", rr.Code)
	}
}
