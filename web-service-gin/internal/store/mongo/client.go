package mongostore

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

func Connect(ctx context.Context, uri string) (*mongo.Client, error) {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	client, err := mongo.Connect(options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		_ = client.Disconnect(ctx)
		return nil, err
	}

	return client, nil
}

func EnsureNoteIndexes(ctx context.Context, database *mongo.Database, collectionName string) error {
	collection := database.Collection(collectionName)
	_, err := collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "userId", Value: 1},
				{Key: "pinned", Value: -1},
				{Key: "updatedAt", Value: -1},
				{Key: "_id", Value: -1},
			},
			Options: options.Index().SetName("notes_user_updated_at_desc"),
		},
		{
			Keys: bson.D{
				{Key: "userId", Value: 1},
				{Key: "_id", Value: 1},
			},
			Options: options.Index().SetName("notes_user_id_lookup"),
		},
	})

	return err
}
