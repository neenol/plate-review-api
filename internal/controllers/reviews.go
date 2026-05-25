package controllers

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"github.com/nglambertjr/plate-review-api/internal/domain"
	"github.com/nglambertjr/plate-review-api/internal/middleware"
)

type reviewService interface {
	Create(ctx context.Context, clerkUserID, state, rawPlate, body string) (*domain.Review, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Review, error)
	List(ctx context.Context, state, rawPlate string) ([]domain.Review, error)
	Delete(ctx context.Context, clerkUserID string, reviewID uuid.UUID) error
}

type ReviewController struct {
	reviews reviewService
}

func NewReviewController(reviews reviewService) *ReviewController {
	return &ReviewController{reviews: reviews}
}

type reviewResponse struct {
	ID        string       `json:"id"`
	PlateID   string       `json:"plate_id"`
	Body      string       `json:"body"`
	Score     int64        `json:"score"`
	CreatedAt string       `json:"created_at"`
	Author    authorDetail `json:"author"`
}

type authorDetail struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

func toReviewResponse(r *domain.Review) reviewResponse {
	a := authorDetail{}
	if r.Author != nil {
		a.ID = r.Author.ID.String()
		a.Username = r.Author.Username
	}
	return reviewResponse{
		ID:        r.ID.String(),
		PlateID:   r.PlateID.String(),
		Body:      r.Body,
		Score:     r.Score,
		CreatedAt: r.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		Author:    a,
	}
}

type createReviewRequest struct {
	Body string `json:"body"`
}

func (c *ReviewController) Create(w http.ResponseWriter, r *http.Request) {
	clerkUserID, ok := middleware.ClerkUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	state := urlParam(r, "state")
	plate := urlParam(r, "plate")

	var req createReviewRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "malformed request body")
		return
	}

	rev, err := c.reviews.Create(r.Context(), clerkUserID, state, plate, req.Body)
	if err != nil {
		httpError(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, toReviewResponse(rev))
}

func (c *ReviewController) List(w http.ResponseWriter, r *http.Request) {
	state := urlParam(r, "state")
	plate := urlParam(r, "plate")

	reviews, err := c.reviews.List(r.Context(), state, plate)
	if err != nil {
		httpError(w, r, err)
		return
	}

	resp := make([]reviewResponse, len(reviews))
	for i, rev := range reviews {
		rev := rev
		resp[i] = toReviewResponse(&rev)
	}
	writeJSON(w, http.StatusOK, resp)
}

func (c *ReviewController) Delete(w http.ResponseWriter, r *http.Request) {
	clerkUserID, ok := middleware.ClerkUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	id, err := urlParamUUID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid review ID")
		return
	}

	if err = c.reviews.Delete(r.Context(), clerkUserID, id); err != nil {
		httpError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
