package notes

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

const MaxLimit = 100

var (
	ErrInvalidID    = errors.New("invalid note id")
	ErrInvalidInput = errors.New("at least one note is required")
	ErrNotFound     = errors.New("note not found")
)

type Note struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Title     string        `bson:"title" json:"title"`
	Content   string        `bson:"content" json:"content"`
	CreatedAt time.Time     `bson:"createdAt,omitempty" json:"createdAt,omitempty"`
	UpdatedAt time.Time     `bson:"updatedAt,omitempty" json:"updatedAt,omitempty"`
}

type Page struct {
	Total int64  `json:"total"`
	Notes []Note `json:"notes"`
	Page  int    `json:"page"`
	Limit int    `json:"limit"`
}

type Pagination struct {
	Page  int
	Limit int
}

type Repository interface {
	List(ctx context.Context, pagination Pagination) ([]Note, int64, error)
	CreateMany(ctx context.Context, newNotes []Note) ([]Note, error)
	FindByID(ctx context.Context, id bson.ObjectID) (Note, error)
}
