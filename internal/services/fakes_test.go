package services_test

import (
	"context"
	"sync"

	"github.com/google/uuid"

	"github.com/nglambertjr/plate-review-api/internal/domain"
)

// fakeUserRepo is an in-memory implementation of services.UserRepository.
type fakeUserRepo struct {
	mu    sync.Mutex
	byID  map[uuid.UUID]*domain.User
	byCID map[string]*domain.User
	byUN  map[string]*domain.User
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{
		byID:  make(map[uuid.UUID]*domain.User),
		byCID: make(map[string]*domain.User),
		byUN:  make(map[string]*domain.User),
	}
}

func (f *fakeUserRepo) seed(u *domain.User) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.byID[u.ID] = u
	f.byCID[u.ClerkUserID] = u
	f.byUN[u.Username] = u
}

func (f *fakeUserRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	u, ok := f.byID[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return u, nil
}

func (f *fakeUserRepo) GetByClerkID(_ context.Context, clerkUserID string) (*domain.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	u, ok := f.byCID[clerkUserID]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return u, nil
}

func (f *fakeUserRepo) GetByUsername(_ context.Context, username string) (*domain.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	u, ok := f.byUN[username]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return u, nil
}

func (f *fakeUserRepo) Create(_ context.Context, clerkUserID, username string) (*domain.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, dup := f.byCID[clerkUserID]; dup {
		return nil, domain.ErrConflict
	}
	if _, dup := f.byUN[username]; dup {
		return nil, domain.ErrConflict
	}
	u := &domain.User{ID: uuid.New(), ClerkUserID: clerkUserID, Username: username}
	f.byID[u.ID] = u
	f.byCID[clerkUserID] = u
	f.byUN[username] = u
	return u, nil
}

func (f *fakeUserRepo) UpdateUsername(_ context.Context, id uuid.UUID, username string) (*domain.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	u, ok := f.byID[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	if _, dup := f.byUN[username]; dup {
		return nil, domain.ErrConflict
	}
	delete(f.byUN, u.Username)
	u.Username = username
	f.byUN[username] = u
	return u, nil
}

// fakePlateRepo is an in-memory PlateRepository.
type fakePlateRepo struct {
	mu      sync.Mutex
	byHash  map[string]*domain.Plate
}

func newFakePlateRepo() *fakePlateRepo {
	return &fakePlateRepo{byHash: make(map[string]*domain.Plate)}
}

func (f *fakePlateRepo) GetByHash(_ context.Context, hash string) (*domain.Plate, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	p, ok := f.byHash[hash]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return p, nil
}

func (f *fakePlateRepo) Create(_ context.Context, hash, state string) (*domain.Plate, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, dup := f.byHash[hash]; dup {
		return nil, domain.ErrConflict
	}
	p := &domain.Plate{ID: uuid.New(), PlateHash: hash, State: state}
	f.byHash[hash] = p
	return p, nil
}

// fakeReviewRepo is an in-memory ReviewRepository.
type fakeReviewRepo struct {
	mu      sync.Mutex
	reviews map[uuid.UUID]*domain.Review
}

func newFakeReviewRepo() *fakeReviewRepo {
	return &fakeReviewRepo{reviews: make(map[uuid.UUID]*domain.Review)}
}

func (f *fakeReviewRepo) Create(_ context.Context, plateID, userID uuid.UUID, body string) (*domain.Review, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	r := &domain.Review{ID: uuid.New(), PlateID: plateID, UserID: userID, Body: body}
	f.reviews[r.ID] = r
	return r, nil
}

func (f *fakeReviewRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Review, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := f.reviews[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return r, nil
}

func (f *fakeReviewRepo) ListByPlate(_ context.Context, plateID uuid.UUID) ([]domain.Review, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []domain.Review
	for _, r := range f.reviews {
		if r.PlateID == plateID {
			out = append(out, *r)
		}
	}
	return out, nil
}

func (f *fakeReviewRepo) Delete(_ context.Context, id uuid.UUID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.reviews, id)
	return nil
}

func (f *fakeReviewRepo) GetOwnerID(_ context.Context, id uuid.UUID) (uuid.UUID, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := f.reviews[id]
	if !ok {
		return uuid.UUID{}, domain.ErrNotFound
	}
	return r.UserID, nil
}

// fakeReplyRepo is an in-memory ReplyRepository.
type fakeReplyRepo struct {
	mu      sync.Mutex
	replies map[uuid.UUID]*domain.Reply
}

func newFakeReplyRepo() *fakeReplyRepo {
	return &fakeReplyRepo{replies: make(map[uuid.UUID]*domain.Reply)}
}

func (f *fakeReplyRepo) Create(_ context.Context, reviewID uuid.UUID, parentReplyID *uuid.UUID, userID uuid.UUID, body string) (*domain.Reply, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	r := &domain.Reply{ID: uuid.New(), ReviewID: reviewID, ParentReplyID: parentReplyID, UserID: userID, Body: body}
	f.replies[r.ID] = r
	return r, nil
}

func (f *fakeReplyRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Reply, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := f.replies[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return r, nil
}

func (f *fakeReplyRepo) ListByReview(_ context.Context, reviewID uuid.UUID) ([]domain.Reply, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []domain.Reply
	for _, r := range f.replies {
		if r.ReviewID == reviewID && r.ParentReplyID == nil {
			out = append(out, *r)
		}
	}
	return out, nil
}

func (f *fakeReplyRepo) ListByParent(_ context.Context, parentID uuid.UUID) ([]domain.Reply, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []domain.Reply
	for _, r := range f.replies {
		if r.ParentReplyID != nil && *r.ParentReplyID == parentID {
			out = append(out, *r)
		}
	}
	return out, nil
}

func (f *fakeReplyRepo) Delete(_ context.Context, id uuid.UUID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.replies, id)
	return nil
}

func (f *fakeReplyRepo) GetOwnerID(_ context.Context, id uuid.UUID) (uuid.UUID, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := f.replies[id]
	if !ok {
		return uuid.UUID{}, domain.ErrNotFound
	}
	return r.UserID, nil
}

// fakeVoteRepo is an in-memory VoteRepository.
type fakeVoteRepo struct {
	mu    sync.Mutex
	votes map[string]*domain.Vote
}

func newFakeVoteRepo() *fakeVoteRepo {
	return &fakeVoteRepo{votes: make(map[string]*domain.Vote)}
}

func voteKey(userID uuid.UUID, targetType string, targetID uuid.UUID) string {
	return userID.String() + "|" + targetType + "|" + targetID.String()
}

func (f *fakeVoteRepo) Create(_ context.Context, userID uuid.UUID, targetType string, targetID uuid.UUID, value int32) (*domain.Vote, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	k := voteKey(userID, targetType, targetID)
	if _, dup := f.votes[k]; dup {
		return nil, domain.ErrConflict
	}
	v := &domain.Vote{ID: uuid.New(), UserID: userID, TargetType: targetType, TargetID: targetID, Value: value}
	f.votes[k] = v
	return v, nil
}

func (f *fakeVoteRepo) GetByUserAndTarget(_ context.Context, userID uuid.UUID, targetType string, targetID uuid.UUID) (*domain.Vote, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	v, ok := f.votes[voteKey(userID, targetType, targetID)]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return v, nil
}

func (f *fakeVoteRepo) Delete(_ context.Context, userID uuid.UUID, targetType string, targetID uuid.UUID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.votes, voteKey(userID, targetType, targetID))
	return nil
}
