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

func (repository *NotesRepository) List(ctx context.Context, userID string, pagination notes.Pagination) ([]notes.Note, int64, error) {
	ctx, cancel := repository.contextWithTimeout(ctx)
	defer cancel()

	filter := listFilter(userID)
	findOptions := options.Find().
		SetLimit(int64(pagination.Limit)).
		SetSkip(int64((pagination.Page - 1) * pagination.Limit)).
		SetSort(listSort())

	cursor, err := repository.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer func() {
		_ = cursor.Close(ctx)
	}()

	total, err := repository.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	storedNotes := make([]notes.Note, 0)
	if err := cursor.All(ctx, &storedNotes); err != nil {
		return nil, 0, err
	}

	return storedNotes, total, nil
}

func (repository *NotesRepository) Create(ctx context.Context, newNote notes.Note) (notes.Note, error) {
	ctx, cancel := repository.contextWithTimeout(ctx)
	defer cancel()

	one, err := repository.collection.InsertOne(ctx, newNote)
	if err != nil {
		return notes.Note{}, err
	}

	var storedNote notes.Note
	if err := repository.collection.FindOne(ctx, bson.M{"_id": one.InsertedID}).Decode(&storedNote); err != nil {
		return notes.Note{}, err
	}

	return storedNote, nil
}

func (repository *NotesRepository) FindByID(ctx context.Context, userID string, id bson.ObjectID) (notes.Note, error) {
	ctx, cancel := repository.contextWithTimeout(ctx)
	defer cancel()

	var result notes.Note
	err := repository.collection.FindOne(ctx, ownedNoteFilter(userID, id)).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return notes.Note{}, notes.ErrNotFound
		}

		return notes.Note{}, err
	}

	return result, nil
}

func (repository *NotesRepository) Update(ctx context.Context, userID string, id bson.ObjectID, update notes.NoteUpdate) (notes.Note, error) {
	ctx, cancel := repository.contextWithTimeout(ctx)
	defer cancel()

	var result notes.Note
	err := repository.collection.FindOneAndUpdate(
		ctx,
		ownedNoteFilter(userID, id),
		bson.M{"$set": updateSet(update)},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return notes.Note{}, notes.ErrNotFound
		}

		return notes.Note{}, err
	}

	return result, nil
}

func (repository *NotesRepository) Delete(ctx context.Context, userID string, id bson.ObjectID) error {
	ctx, cancel := repository.contextWithTimeout(ctx)
	defer cancel()

	result, err := repository.collection.DeleteOne(ctx, ownedNoteFilter(userID, id))
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return notes.ErrNotFound
	}

	return nil
}

func (repository *NotesRepository) contextWithTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if repository.operationTimeout <= 0 {
		return context.WithCancel(ctx)
	}

	return context.WithTimeout(ctx, repository.operationTimeout)
}

func ownedNoteFilter(userID string, id bson.ObjectID) bson.M {
	return bson.M{
		"_id":    id,
		"userId": userID,
	}
}

func listFilter(userID string) bson.M {
	return bson.M{"userId": userID}
}

func listSort() bson.D {
	return bson.D{
		{Key: "pinned", Value: -1},
		{Key: "updatedAt", Value: -1},
		{Key: "_id", Value: -1},
	}
}

func updateSet(update notes.NoteUpdate) bson.M {
	set := bson.M{"updatedAt": update.UpdatedAt}
	if update.Title != nil {
		set["title"] = *update.Title
	}
	if update.Content != nil {
		set["content"] = *update.Content
	}
	if update.Pinned != nil {
		set["pinned"] = *update.Pinned
	}

	return set
}
