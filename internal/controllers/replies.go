package controllers

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"github.com/nglambertjr/plate-review-api/internal/domain"
	"github.com/nglambertjr/plate-review-api/internal/middleware"
)

type replyService interface {
	CreateOnReview(ctx context.Context, clerkUserID string, reviewID uuid.UUID, body string) (*domain.Reply, error)
	CreateOnReply(ctx context.Context, clerkUserID string, parentReplyID uuid.UUID, body string) (*domain.Reply, error)
	ListByReview(ctx context.Context, reviewID uuid.UUID) ([]domain.Reply, error)
	ListByParent(ctx context.Context, parentReplyID uuid.UUID) ([]domain.Reply, error)
	Delete(ctx context.Context, clerkUserID string, replyID uuid.UUID) error
}

type ReplyController struct {
	replies replyService
}

func NewReplyController(replies replyService) *ReplyController {
	return &ReplyController{replies: replies}
}

type replyResponse struct {
	ID            string       `json:"id"`
	ReviewID      string       `json:"review_id"`
	ParentReplyID *string      `json:"parent_reply_id,omitempty"`
	Body          string       `json:"body"`
	Score         int64        `json:"score"`
	CreatedAt     string       `json:"created_at"`
	Author        authorDetail `json:"author"`
}

func toReplyResponse(r *domain.Reply) replyResponse {
	a := authorDetail{}
	if r.Author != nil {
		a.ID = r.Author.ID.String()
		a.Username = r.Author.Username
	}
	resp := replyResponse{
		ID:        r.ID.String(),
		ReviewID:  r.ReviewID.String(),
		Body:      r.Body,
		Score:     r.Score,
		CreatedAt: r.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		Author:    a,
	}
	if r.ParentReplyID != nil {
		s := r.ParentReplyID.String()
		resp.ParentReplyID = &s
	}
	return resp
}

type createReplyRequest struct {
	Body string `json:"body"`
}

// CreateOnReview handles POST /reviews/{id}/replies
func (c *ReplyController) CreateOnReview(w http.ResponseWriter, r *http.Request) {
	clerkUserID, ok := middleware.ClerkUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	reviewID, err := urlParamUUID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid review ID")
		return
	}

	var req createReplyRequest
	if err = decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "malformed request body")
		return
	}

	reply, err := c.replies.CreateOnReview(r.Context(), clerkUserID, reviewID, req.Body)
	if err != nil {
		httpError(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, toReplyResponse(reply))
}

// CreateOnReply handles POST /replies/{id}/replies
func (c *ReplyController) CreateOnReply(w http.ResponseWriter, r *http.Request) {
	clerkUserID, ok := middleware.ClerkUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	parentID, err := urlParamUUID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid reply ID")
		return
	}

	var req createReplyRequest
	if err = decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "malformed request body")
		return
	}

	reply, err := c.replies.CreateOnReply(r.Context(), clerkUserID, parentID, req.Body)
	if err != nil {
		httpError(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, toReplyResponse(reply))
}

// ListByReview handles GET /reviews/{id}/replies
func (c *ReplyController) ListByReview(w http.ResponseWriter, r *http.Request) {
	reviewID, err := urlParamUUID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid review ID")
		return
	}

	replies, err := c.replies.ListByReview(r.Context(), reviewID)
	if err != nil {
		httpError(w, r, err)
		return
	}

	resp := make([]replyResponse, len(replies))
	for i, rep := range replies {
		rep := rep
		resp[i] = toReplyResponse(&rep)
	}
	writeJSON(w, http.StatusOK, resp)
}

// ListByParent handles GET /replies/{id}/replies
func (c *ReplyController) ListByParent(w http.ResponseWriter, r *http.Request) {
	parentID, err := urlParamUUID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid reply ID")
		return
	}

	replies, err := c.replies.ListByParent(r.Context(), parentID)
	if err != nil {
		httpError(w, r, err)
		return
	}

	resp := make([]replyResponse, len(replies))
	for i, rep := range replies {
		rep := rep
		resp[i] = toReplyResponse(&rep)
	}
	writeJSON(w, http.StatusOK, resp)
}

// Delete handles DELETE /replies/{id}
func (c *ReplyController) Delete(w http.ResponseWriter, r *http.Request) {
	clerkUserID, ok := middleware.ClerkUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	id, err := urlParamUUID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid reply ID")
		return
	}

	if err = c.replies.Delete(r.Context(), clerkUserID, id); err != nil {
		httpError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
