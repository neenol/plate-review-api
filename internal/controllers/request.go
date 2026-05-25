package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func decodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("malformed JSON: %w", errors.New("invalid_input"))
	}
	return nil
}

func urlParamUUID(r *http.Request, key string) (uuid.UUID, error) {
	raw := chi.URLParam(r, key)
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("invalid UUID for %q: %w", key, errors.New("invalid_input"))
	}
	return id, nil
}

func urlParam(r *http.Request, key string) string {
	return chi.URLParam(r, key)
}
