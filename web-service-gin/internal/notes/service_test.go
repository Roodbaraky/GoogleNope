package notes

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestServiceListClampsLimit(t *testing.T) {
	repository := &serviceFakeRepository{}
	service := NewService(repository)

	page, err := service.List(context.Background(), 2, MaxLimit+50)
	if err != nil {
		t.Fatalf("expected list to succeed: %v", err)
	}

	if repository.pagination.Page != 2 {
		t.Fatalf("expected page 2, got %d", repository.pagination.Page)
	}

	if repository.pagination.Limit != MaxLimit {
		t.Fatalf("expected limit to be clamped to %d, got %d", MaxLimit, repository.pagination.Limit)
	}

	if page.Limit != MaxLimit {
		t.Fatalf("expected response limit to be %d, got %d", MaxLimit, page.Limit)
	}
}

func TestServiceCreateManyRejectsEmptyInput(t *testing.T) {
	service := NewService(&serviceFakeRepository{})

	_, err := service.CreateMany(context.Background(), nil)
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestServiceCreateManySanitizesAndTimestampsNotes(t *testing.T) {
	repository := &serviceFakeRepository{}
	service := NewService(repository)
	fixedNow := time.Date(2026, 5, 27, 12, 0, 0, 0, time.FixedZone("BST", 3600))
	service.now = func() time.Time {
		return fixedNow
	}

	clientProvidedID := bson.NewObjectID()
	createdAt := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	_, err := service.CreateMany(context.Background(), []Note{
		{
			ID:        clientProvidedID,
			Title:     "Title",
			Content:   "Content",
			CreatedAt: createdAt,
		},
	})
	if err != nil {
		t.Fatalf("expected create to succeed: %v", err)
	}

	if len(repository.created) != 1 {
		t.Fatalf("expected one created note, got %d", len(repository.created))
	}

	created := repository.created[0]
	if created.ID != (bson.ObjectID{}) {
		t.Fatalf("expected service to strip client-provided ID, got %s", created.ID.Hex())
	}

	if !created.CreatedAt.Equal(createdAt) {
		t.Fatalf("expected service to preserve CreatedAt, got %s", created.CreatedAt)
	}

	if !created.UpdatedAt.Equal(fixedNow.UTC()) {
		t.Fatalf("expected service to set UpdatedAt to %s, got %s", fixedNow.UTC(), created.UpdatedAt)
	}
}

func TestServiceFindByIDRejectsInvalidID(t *testing.T) {
	repository := &serviceFakeRepository{}
	service := NewService(repository)

	_, err := service.FindByID(context.Background(), "not-an-object-id")
	if !errors.Is(err, ErrInvalidID) {
		t.Fatalf("expected ErrInvalidID, got %v", err)
	}

	if repository.findCalled {
		t.Fatal("expected invalid ID to be rejected before repository call")
	}
}

type serviceFakeRepository struct {
	pagination Pagination
	created    []Note
	findCalled bool
}

func (repository *serviceFakeRepository) List(_ context.Context, pagination Pagination) ([]Note, int64, error) {
	repository.pagination = pagination
	return []Note{}, 0, nil
}

func (repository *serviceFakeRepository) CreateMany(_ context.Context, newNotes []Note) ([]Note, error) {
	repository.created = append(repository.created, newNotes...)
	return newNotes, nil
}

func (repository *serviceFakeRepository) FindByID(_ context.Context, _ bson.ObjectID) (Note, error) {
	repository.findCalled = true
	return Note{}, ErrNotFound
}
