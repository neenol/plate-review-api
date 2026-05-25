package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/nglambertjr/plate-review-api/internal/db/generated"
	"github.com/nglambertjr/plate-review-api/internal/domain"
)

type ReplyRepository struct {
	q *db.Queries
}

func NewReplyRepository(dbtx db.DBTX) *ReplyRepository {
	return &ReplyRepository{q: db.New(dbtx)}
}

func (r *ReplyRepository) Create(ctx context.Context, reviewID uuid.UUID, parentReplyID *uuid.UUID, userID uuid.UUID, body string) (*domain.Reply, error) {
	row, err := r.q.CreateReply(ctx, db.CreateReplyParams{
		ReviewID:      reviewID,
		ParentReplyID: uuidPtrToPgUUID(parentReplyID),
		UserID:        userID,
		Body:          body,
	})
	if err != nil {
		return nil, fmt.Errorf("reply repository Create: %w", err)
	}
	return &domain.Reply{
		ID:            row.ID,
		ReviewID:      row.ReviewID,
		ParentReplyID: pgUUIDToPtr(row.ParentReplyID),
		UserID:        row.UserID,
		Body:          row.Body,
		CreatedAt:     row.CreatedAt.Time,
	}, nil
}

func (r *ReplyRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Reply, error) {
	row, err := r.q.GetReplyByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("reply repository GetByID: %w", err)
	}
	return mapReplyRow(row), nil
}

func (r *ReplyRepository) ListByReview(ctx context.Context, reviewID uuid.UUID) ([]domain.Reply, error) {
	rows, err := r.q.ListRepliesByReview(ctx, reviewID)
	if err != nil {
		return nil, fmt.Errorf("reply repository ListByReview: %w", err)
	}
	out := make([]domain.Reply, len(rows))
	for i, row := range rows {
		out[i] = *mapListReplyByReviewRow(row)
	}
	return out, nil
}

func (r *ReplyRepository) ListByParent(ctx context.Context, parentID uuid.UUID) ([]domain.Reply, error) {
	rows, err := r.q.ListRepliesByParent(ctx, pgtype.UUID{Bytes: parentID, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("reply repository ListByParent: %w", err)
	}
	out := make([]domain.Reply, len(rows))
	for i, row := range rows {
		out[i] = *mapListReplyByParentRow(row)
	}
	return out, nil
}

func (r *ReplyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.q.DeleteReply(ctx, id); err != nil {
		return fmt.Errorf("reply repository Delete: %w", err)
	}
	return nil
}

func (r *ReplyRepository) GetOwnerID(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	ownerID, err := r.q.GetReplyOwnerID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.UUID{}, domain.ErrNotFound
		}
		return uuid.UUID{}, fmt.Errorf("reply repository GetOwnerID: %w", err)
	}
	return ownerID, nil
}

func mapReplyRow(row db.GetReplyByIDRow) *domain.Reply {
	return &domain.Reply{
		ID:            row.ID,
		ReviewID:      row.ReviewID,
		ParentReplyID: pgUUIDToPtr(row.ParentReplyID),
		UserID:        row.UserID,
		Body:          row.Body,
		Score:         row.Score,
		CreatedAt:     row.CreatedAt.Time,
		Author:        &domain.User{ID: row.UserID, Username: row.AuthorUsername},
	}
}

func mapListReplyByReviewRow(row db.ListRepliesByReviewRow) *domain.Reply {
	return &domain.Reply{
		ID:            row.ID,
		ReviewID:      row.ReviewID,
		ParentReplyID: pgUUIDToPtr(row.ParentReplyID),
		UserID:        row.UserID,
		Body:          row.Body,
		Score:         row.Score,
		CreatedAt:     row.CreatedAt.Time,
		Author:        &domain.User{ID: row.UserID, Username: row.AuthorUsername},
	}
}

func mapListReplyByParentRow(row db.ListRepliesByParentRow) *domain.Reply {
	return &domain.Reply{
		ID:            row.ID,
		ReviewID:      row.ReviewID,
		ParentReplyID: pgUUIDToPtr(row.ParentReplyID),
		UserID:        row.UserID,
		Body:          row.Body,
		Score:         row.Score,
		CreatedAt:     row.CreatedAt.Time,
		Author:        &domain.User{ID: row.UserID, Username: row.AuthorUsername},
	}
}

func pgUUIDToPtr(pu pgtype.UUID) *uuid.UUID {
	if !pu.Valid {
		return nil
	}
	id := uuid.UUID(pu.Bytes)
	return &id
}

func uuidPtrToPgUUID(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: *id, Valid: true}
}
