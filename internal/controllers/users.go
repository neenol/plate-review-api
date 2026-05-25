package controllers

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"github.com/nglambertjr/plate-review-api/internal/domain"
	"github.com/nglambertjr/plate-review-api/internal/middleware"
)

type userService interface {
	Create(ctx context.Context, clerkUserID, username string) (*domain.User, error)
	GetByClerkID(ctx context.Context, clerkUserID string) (*domain.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	UpdateUsername(ctx context.Context, clerkUserID, newUsername string) (*domain.User, error)
}

type UserController struct {
	users userService
}

func NewUserController(users userService) *UserController {
	return &UserController{users: users}
}

type createUserRequest struct {
	Username string `json:"username"`
}

type userResponse struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	ClerkUserID string `json:"clerk_user_id"`
	CreatedAt   string `json:"created_at"`
}

func toUserResponse(u *domain.User) userResponse {
	return userResponse{
		ID:          u.ID.String(),
		Username:    u.Username,
		ClerkUserID: u.ClerkUserID,
		CreatedAt:   u.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
}

func (c *UserController) Create(w http.ResponseWriter, r *http.Request) {
	clerkUserID, ok := middleware.ClerkUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	var req createUserRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "malformed request body")
		return
	}

	u, err := c.users.Create(r.Context(), clerkUserID, req.Username)
	if err != nil {
		httpError(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, toUserResponse(u))
}

func (c *UserController) Me(w http.ResponseWriter, r *http.Request) {
	clerkUserID, ok := middleware.ClerkUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	u, err := c.users.GetByClerkID(r.Context(), clerkUserID)
	if err != nil {
		httpError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, toUserResponse(u))
}

type updateUsernameRequest struct {
	Username string `json:"username"`
}

func (c *UserController) UpdateMe(w http.ResponseWriter, r *http.Request) {
	clerkUserID, ok := middleware.ClerkUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	var req updateUsernameRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "malformed request body")
		return
	}

	u, err := c.users.UpdateUsername(r.Context(), clerkUserID, req.Username)
	if err != nil {
		httpError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, toUserResponse(u))
}
