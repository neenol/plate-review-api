package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/nglambertjr/plate-review-api/internal/domain"
	"github.com/nglambertjr/plate-review-api/internal/plate"
)

type PlateService struct {
	plates PlateRepository
	pepper string
}

func NewPlateService(plates PlateRepository, pepper string) *PlateService {
	return &PlateService{plates: plates, pepper: pepper}
}

// GetOrCreate normalizes the plate number, hashes it, and returns the plate
// record—creating it if it doesn't exist yet.
func (s *PlateService) GetOrCreate(ctx context.Context, state, rawPlate string) (*domain.Plate, error) {
	_, hash, err := plate.NormalizeAndHash(state, rawPlate, s.pepper)
	if err != nil {
		return nil, err
	}

	p, err := s.plates.GetByHash(ctx, hash)
	if err == nil {
		return p, nil
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return nil, fmt.Errorf("plate service GetOrCreate lookup: %w", err)
	}

	// Race: two requests for the same plate may both attempt to create it.
	// If the second insert gets a conflict error, fall back to a read.
	upperState := state
	p, err = s.plates.Create(ctx, hash, upperState)
	if err != nil {
		if errors.Is(err, domain.ErrConflict) {
			p, err = s.plates.GetByHash(ctx, hash)
			if err != nil {
				return nil, fmt.Errorf("plate service GetOrCreate after conflict: %w", err)
			}
			return p, nil
		}
		return nil, fmt.Errorf("plate service GetOrCreate create: %w", err)
	}
	return p, nil
}
