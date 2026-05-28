package mongostore

import (
	"context"
	"errors"
	"time"

	"example/web-service-gin/internal/notes"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type NotesRepository struct {
	collection       *mongo.Collection
	operationTimeout time.Duration
}

func NewNotesRepository(database *mongo.Database, collectionName string, operationTimeout time.Duration) *NotesRepository {
	return &NotesRepository{
		collection:       database.Collection(collectionName),
		operationTimeout: operationTimeout,
	}
}

func (repository *NotesRepository) List(ctx context.Context, pagination notes.Pagination) ([]notes.Note, int64, error) {
	ctx, cancel := repository.contextWithTimeout(ctx)
	defer cancel()

	findOptions := options.Find().
		SetLimit(int64(pagination.Limit)).
		SetSkip(int64((pagination.Page - 1) * pagination.Limit)).
		SetSort(bson.D{
			{Key: "updatedAt", Value: -1},
			{Key: "_id", Value: -1},
		})

	cursor, err := repository.collection.Find(ctx, bson.D{}, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer func() {
		_ = cursor.Close(ctx)
	}()

	total, err := repository.collection.CountDocuments(ctx, bson.D{})
	if err != nil {
		return nil, 0, err
	}

	storedNotes := make([]notes.Note, 0)
	if err := cursor.All(ctx, &storedNotes); err != nil {
		return nil, 0, err
	}

	return storedNotes, total, nil
}

func (repository *NotesRepository) CreateMany(ctx context.Context, newNotes []notes.Note) ([]notes.Note, error) {
	ctx, cancel := repository.contextWithTimeout(ctx)
	defer cancel()

	docs := make([]interface{}, len(newNotes))
	for i, note := range newNotes {
		docs[i] = note
	}

	many, err := repository.collection.InsertMany(ctx, docs)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"_id": bson.M{"$in": many.InsertedIDs}}
	cursor, err := repository.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = cursor.Close(ctx)
	}()

	storedNotes := make([]notes.Note, 0, len(newNotes))
	if err = cursor.All(ctx, &storedNotes); err != nil {
		return nil, err
	}

	return storedNotes, nil
}

func (repository *NotesRepository) FindByID(ctx context.Context, id bson.ObjectID) (notes.Note, error) {
	ctx, cancel := repository.contextWithTimeout(ctx)
	defer cancel()

	var result notes.Note
	err := repository.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return notes.Note{}, notes.ErrNotFound
		}

		return notes.Note{}, err
	}

	return result, nil
}

func (repository *NotesRepository) contextWithTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if repository.operationTimeout <= 0 {
		return context.WithCancel(ctx)
	}

	return context.WithTimeout(ctx, repository.operationTimeout)
}
