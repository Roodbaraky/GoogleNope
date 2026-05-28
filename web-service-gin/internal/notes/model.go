package notes

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

const MaxLimit = 100
const MaxTitleLength = 120
const MaxContentLength = 20000

var (
	ErrInvalidID    = errors.New("invalid note id")
	ErrInvalidInput = errors.New("invalid note input")
	ErrNotFound     = errors.New("note not found")
)

type Note struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID    string        `bson:"userId" json:"-"`
	Title     string        `bson:"title" json:"title"`
	Content   string        `bson:"content" json:"content"`
	Pinned    bool          `bson:"pinned" json:"pinned"`
	CreatedAt time.Time     `bson:"createdAt,omitempty" json:"createdAt,omitempty"`
	UpdatedAt time.Time     `bson:"updatedAt,omitempty" json:"updatedAt,omitempty"`
}

type Page struct {
	Total int64  `json:"total"`
	Notes []Note `json:"notes"`
	Page  int    `json:"page"`
	Limit int    `json:"limit"`
	Pages int64  `json:"pages"`
}

type Pagination struct {
	Page  int
	Limit int
}

type CreateNoteInput struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Pinned  bool   `json:"pinned"`
}

type UpdateNoteInput struct {
	Title   *string `json:"title"`
	Content *string `json:"content"`
	Pinned  *bool   `json:"pinned"`
}

type NoteUpdate struct {
	Title     *string
	Content   *string
	Pinned    *bool
	UpdatedAt time.Time
}

type Repository interface {
	List(ctx context.Context, userID string, pagination Pagination) ([]Note, int64, error)
	Create(ctx context.Context, newNote Note) (Note, error)
	FindByID(ctx context.Context, userID string, id bson.ObjectID) (Note, error)
	Update(ctx context.Context, userID string, id bson.ObjectID, update NoteUpdate) (Note, error)
	Delete(ctx context.Context, userID string, id bson.ObjectID) error
}
