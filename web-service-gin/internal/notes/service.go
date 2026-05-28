package notes

import (
	"context"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Service struct {
	repository Repository
	now        func() time.Time
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
		now:        time.Now,
	}
}

func (service *Service) List(ctx context.Context, userID string, page int, limit int) (Page, error) {
	if limit > MaxLimit {
		limit = MaxLimit
	}

	storedNotes, total, err := service.repository.List(ctx, userID, Pagination{
		Page:  page,
		Limit: limit,
	})
	if err != nil {
		return Page{}, err
	}

	return Page{
		Total: total,
		Notes: storedNotes,
		Page:  page,
		Limit: limit,
		Pages: pageCount(total, limit),
	}, nil
}

func (service *Service) Create(ctx context.Context, userID string, input CreateNoteInput) (Note, error) {
	title := strings.TrimSpace(input.Title)
	content := strings.TrimSpace(input.Content)
	if !validNoteText(title, content) {
		return Note{}, ErrInvalidInput
	}

	now := service.now().UTC()
	newNote := Note{
		UserID:    userID,
		Title:     title,
		Content:   content,
		Pinned:    input.Pinned,
		CreatedAt: now,
		UpdatedAt: now,
	}

	return service.repository.Create(ctx, newNote)
}

func (service *Service) FindByID(ctx context.Context, userID string, id string) (Note, error) {
	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return Note{}, ErrInvalidID
	}

	return service.repository.FindByID(ctx, userID, objID)
}

func (service *Service) Update(ctx context.Context, userID string, id string, input UpdateNoteInput) (Note, error) {
	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return Note{}, ErrInvalidID
	}

	if input.Title == nil && input.Content == nil && input.Pinned == nil {
		return Note{}, ErrInvalidInput
	}

	update := NoteUpdate{
		Pinned:    input.Pinned,
		UpdatedAt: service.now().UTC(),
	}

	if input.Title != nil {
		title := strings.TrimSpace(*input.Title)
		if len(title) > MaxTitleLength {
			return Note{}, ErrInvalidInput
		}

		update.Title = &title
	}

	if input.Content != nil {
		content := strings.TrimSpace(*input.Content)
		if len(content) > MaxContentLength {
			return Note{}, ErrInvalidInput
		}

		update.Content = &content
	}

	if update.Title != nil && update.Content != nil && *update.Title == "" && *update.Content == "" {
		return Note{}, ErrInvalidInput
	}

	return service.repository.Update(ctx, userID, objID, update)
}

func (service *Service) Delete(ctx context.Context, userID string, id string) error {
	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return ErrInvalidID
	}

	return service.repository.Delete(ctx, userID, objID)
}

func validNoteText(title string, content string) bool {
	if title == "" && content == "" {
		return false
	}

	return len(title) <= MaxTitleLength && len(content) <= MaxContentLength
}

func pageCount(total int64, limit int) int64 {
	if total == 0 {
		return 0
	}

	pages := total / int64(limit)
	if total%int64(limit) != 0 {
		pages++
	}

	return pages
}
