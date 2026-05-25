package controllers

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"github.com/nglambertjr/plate-review-api/internal/domain"
	"github.com/nglambertjr/plate-review-api/internal/middleware"
)

type voteService interface {
	CastVote(ctx context.Context, clerkUserID, targetType string, targetID uuid.UUID) (*domain.Vote, error)
	RetractVote(ctx context.Context, clerkUserID, targetType string, targetID uuid.UUID) error
}

type VoteController struct {
	votes voteService
}

func NewVoteController(votes voteService) *VoteController {
	return &VoteController{votes: votes}
}

type voteResponse struct {
	ID         string `json:"id"`
	TargetType string `json:"target_type"`
	TargetID   string `json:"target_id"`
	Value      int32  `json:"value"`
	CreatedAt  string `json:"created_at"`
}

func toVoteResponse(v *domain.Vote) voteResponse {
	return voteResponse{
		ID:         v.ID.String(),
		TargetType: v.TargetType,
		TargetID:   v.TargetID.String(),
		Value:      v.Value,
		CreatedAt:  v.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
}

// CastReviewVote handles POST /reviews/{id}/votes
func (c *VoteController) CastReviewVote(w http.ResponseWriter, r *http.Request) {
	c.castVote(w, r, domain.VoteTargetReview)
}

// RetractReviewVote handles DELETE /reviews/{id}/votes
func (c *VoteController) RetractReviewVote(w http.ResponseWriter, r *http.Request) {
	c.retractVote(w, r, domain.VoteTargetReview)
}

// CastReplyVote handles POST /replies/{id}/votes
func (c *VoteController) CastReplyVote(w http.ResponseWriter, r *http.Request) {
	c.castVote(w, r, domain.VoteTargetReply)
}

// RetractReplyVote handles DELETE /replies/{id}/votes
func (c *VoteController) RetractReplyVote(w http.ResponseWriter, r *http.Request) {
	c.retractVote(w, r, domain.VoteTargetReply)
}

func (c *VoteController) castVote(w http.ResponseWriter, r *http.Request, targetType string) {
	clerkUserID, ok := middleware.ClerkUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	targetID, err := urlParamUUID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid ID")
		return
	}

	vote, err := c.votes.CastVote(r.Context(), clerkUserID, targetType, targetID)
	if err != nil {
		httpError(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, toVoteResponse(vote))
}

func (c *VoteController) retractVote(w http.ResponseWriter, r *http.Request, targetType string) {
	clerkUserID, ok := middleware.ClerkUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	targetID, err := urlParamUUID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid ID")
		return
	}

	if err = c.votes.RetractVote(r.Context(), clerkUserID, targetType, targetID); err != nil {
		httpError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
