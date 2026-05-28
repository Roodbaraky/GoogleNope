package notes

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestServiceListClampsLimit(t *testing.T) {
	repository := &serviceFakeRepository{}
	service := NewService(repository)

	page, err := service.List(context.Background(), "user-1", 2, MaxLimit+50)
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

	if repository.userID != "user-1" {
		t.Fatalf("expected user scoped list, got %q", repository.userID)
	}

	if page.Total != 25 || page.Pages != 1 {
		t.Fatalf("expected total and pages metadata, got total=%d pages=%d", page.Total, page.Pages)
	}
}

func TestServiceListCalculatesPageCount(t *testing.T) {
	repository := &serviceFakeRepository{listTotal: 25}
	service := NewService(repository)

	page, err := service.List(context.Background(), "user-1", 1, 10)
	if err != nil {
		t.Fatalf("expected list to succeed: %v", err)
	}

	if page.Pages != 3 {
		t.Fatalf("expected 3 pages, got %d", page.Pages)
	}
}

func TestServiceCreateRejectsEmptyInput(t *testing.T) {
	service := NewService(&serviceFakeRepository{})

	_, err := service.Create(context.Background(), "user-1", CreateNoteInput{})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestServiceCreateRejectsOverLengthInput(t *testing.T) {
	service := NewService(&serviceFakeRepository{})

	_, err := service.Create(context.Background(), "user-1", CreateNoteInput{
		Title: strings.Repeat("a", MaxTitleLength+1),
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput for long title, got %v", err)
	}

	_, err = service.Create(context.Background(), "user-1", CreateNoteInput{
		Content: strings.Repeat("a", MaxContentLength+1),
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput for long content, got %v", err)
	}
}

func TestServiceCreateSanitizesAndTimestampsNote(t *testing.T) {
	repository := &serviceFakeRepository{}
	service := NewService(repository)
	fixedNow := time.Date(2026, 5, 27, 12, 0, 0, 0, time.FixedZone("BST", 3600))
	service.now = func() time.Time {
		return fixedNow
	}

	_, err := service.Create(context.Background(), "user-1", CreateNoteInput{
		Title:   "  Title  ",
		Content: "  Content  ",
		Pinned:  true,
	})
	if err != nil {
		t.Fatalf("expected create to succeed: %v", err)
	}

	created := repository.created
	if created.UserID != "user-1" {
		t.Fatalf("expected user id to be assigned, got %q", created.UserID)
	}

	if created.Title != "Title" {
		t.Fatalf("expected title to be trimmed, got %q", created.Title)
	}

	if created.Content != "Content" {
		t.Fatalf("expected content to be trimmed, got %q", created.Content)
	}

	if !created.Pinned {
		t.Fatal("expected pinned flag to be preserved")
	}

	if !created.CreatedAt.Equal(fixedNow.UTC()) {
		t.Fatalf("expected CreatedAt to be %s, got %s", fixedNow.UTC(), created.CreatedAt)
	}

	if !created.UpdatedAt.Equal(fixedNow.UTC()) {
		t.Fatalf("expected service to set UpdatedAt to %s, got %s", fixedNow.UTC(), created.UpdatedAt)
	}
}

func TestServiceFindByIDRejectsInvalidID(t *testing.T) {
	repository := &serviceFakeRepository{}
	service := NewService(repository)

	_, err := service.FindByID(context.Background(), "user-1", "not-an-object-id")
	if !errors.Is(err, ErrInvalidID) {
		t.Fatalf("expected ErrInvalidID, got %v", err)
	}

	if repository.findCalled {
		t.Fatal("expected invalid ID to be rejected before repository call")
	}
}

func TestServiceUpdateRejectsEmptyPatch(t *testing.T) {
	service := NewService(&serviceFakeRepository{})

	_, err := service.Update(context.Background(), "user-1", bson.NewObjectID().Hex(), UpdateNoteInput{})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestServiceUpdateRejectsOverLengthInput(t *testing.T) {
	service := NewService(&serviceFakeRepository{})
	id := bson.NewObjectID().Hex()

	title := strings.Repeat("a", MaxTitleLength+1)
	_, err := service.Update(context.Background(), "user-1", id, UpdateNoteInput{Title: &title})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput for long title, got %v", err)
	}

	content := strings.Repeat("a", MaxContentLength+1)
	_, err = service.Update(context.Background(), "user-1", id, UpdateNoteInput{Content: &content})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput for long content, got %v", err)
	}
}

func TestServiceUpdateRejectsClearingTitleAndContentTogether(t *testing.T) {
	service := NewService(&serviceFakeRepository{})
	emptyTitle := " "
	emptyContent := " "

	_, err := service.Update(context.Background(), "user-1", bson.NewObjectID().Hex(), UpdateNoteInput{
		Title:   &emptyTitle,
		Content: &emptyContent,
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestServiceUpdateTrimsTextAndTimestamps(t *testing.T) {
	repository := &serviceFakeRepository{}
	service := NewService(repository)
	fixedNow := time.Date(2026, 5, 27, 12, 0, 0, 0, time.UTC)
	service.now = func() time.Time {
		return fixedNow
	}

	title := "  Updated  "
	pinned := true
	_, err := service.Update(context.Background(), "user-1", bson.NewObjectID().Hex(), UpdateNoteInput{
		Title:  &title,
		Pinned: &pinned,
	})
	if err != nil {
		t.Fatalf("expected update to succeed: %v", err)
	}

	if repository.updated.Title == nil || *repository.updated.Title != "Updated" {
		t.Fatalf("expected trimmed title, got %#v", repository.updated.Title)
	}

	if repository.updated.Pinned == nil || !*repository.updated.Pinned {
		t.Fatal("expected pinned update")
	}

	if !repository.updated.UpdatedAt.Equal(fixedNow) {
		t.Fatalf("expected UpdatedAt to be %s, got %s", fixedNow, repository.updated.UpdatedAt)
	}
}

type serviceFakeRepository struct {
	pagination Pagination
	userID     string
	created    Note
	updated    NoteUpdate
	findCalled bool
	listTotal  int64
}

func (repository *serviceFakeRepository) List(_ context.Context, userID string, pagination Pagination) ([]Note, int64, error) {
	repository.userID = userID
	repository.pagination = pagination
	total := repository.listTotal
	if total == 0 {
		total = 25
	}

	return []Note{}, total, nil
}

func (repository *serviceFakeRepository) Create(_ context.Context, newNote Note) (Note, error) {
	repository.created = newNote
	return newNote, nil
}

func (repository *serviceFakeRepository) FindByID(_ context.Context, _ string, _ bson.ObjectID) (Note, error) {
	repository.findCalled = true
	return Note{}, ErrNotFound
}

func (repository *serviceFakeRepository) Update(_ context.Context, _ string, _ bson.ObjectID, update NoteUpdate) (Note, error) {
	repository.updated = update
	return Note{}, nil
}

func (repository *serviceFakeRepository) Delete(_ context.Context, _ string, _ bson.ObjectID) error {
	return nil
}
