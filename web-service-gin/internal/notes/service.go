package notes

import (
	"context"
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

func (service *Service) List(ctx context.Context, page int, limit int) (Page, error) {
	if limit > MaxLimit {
		limit = MaxLimit
	}

	storedNotes, total, err := service.repository.List(ctx, Pagination{
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
	}, nil
}

func (service *Service) CreateMany(ctx context.Context, newNotes []Note) ([]Note, error) {
	if len(newNotes) == 0 {
		return nil, ErrInvalidInput
	}

	now := service.now().UTC()
	sanitized := make([]Note, len(newNotes))
	for i, note := range newNotes {
		note.ID = bson.ObjectID{}
		if note.CreatedAt.IsZero() {
			note.CreatedAt = now
		}
		note.UpdatedAt = now
		sanitized[i] = note
	}

	return service.repository.CreateMany(ctx, sanitized)
}

func (service *Service) FindByID(ctx context.Context, id string) (Note, error) {
	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return Note{}, ErrInvalidID
	}

	return service.repository.FindByID(ctx, objID)
}
