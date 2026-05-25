package controllers

import (
	"context"
	"net/http"

	"github.com/nglambertjr/plate-review-api/internal/domain"
)

type plateService interface {
	GetOrCreate(ctx context.Context, state, rawPlate string) (*domain.Plate, error)
}

type PlateController struct {
	plates plateService
}

func NewPlateController(plates plateService) *PlateController {
	return &PlateController{plates: plates}
}

type plateResponse struct {
	ID        string `json:"id"`
	State     string `json:"state"`
	CreatedAt string `json:"created_at"`
}

func toPlateResponse(p *domain.Plate) plateResponse {
	return plateResponse{
		ID:        p.ID.String(),
		State:     p.State,
		CreatedAt: p.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
}

// Get resolves (or creates) the canonical plate record and returns its metadata.
func (c *PlateController) Get(w http.ResponseWriter, r *http.Request) {
	state := urlParam(r, "state")
	plate := urlParam(r, "plate")

	p, err := c.plates.GetOrCreate(r.Context(), state, plate)
	if err != nil {
		httpError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, toPlateResponse(p))
}
