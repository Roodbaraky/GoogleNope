package main

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"log"
	"net/http"
	"os"
)

func goDotEnvVariable(key string) string {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	return os.Getenv(key)
}

var client *mongo.Client

func main() {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	uri := goDotEnvVariable("MONGODB_URI")
	var err error
	client, err = mongo.Connect(options.Client().
		ApplyURI(uri).SetServerAPIOptions(serverAPI))
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	router := gin.Default()
	router.GET("/notes", getNotes)
	router.POST("/notes", postNotes)
	router.GET("/notes/:id", getNoteById)

	serverErr := router.Run("localhost:8080")
	if serverErr != nil {
		return
	}
}

type note struct {
	ID      bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Title   string        `json:"title"`
	Content string        `json:"content"`
}

func getNotes(ctx *gin.Context) {
	collection := client.Database("notes").Collection("notes")

	cursor, err := collection.Find(context.TODO(), bson.D{})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer func() {
		if err := cursor.Close(context.TODO()); err != nil {
			log.Printf("Error closing cursor: %v", err)
		}
	}()

	var notes []note
	if err = cursor.All(context.TODO(), &notes); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding notes"})
		return
	}

	ctx.JSON(http.StatusOK, notes)
}

func postNotes(ctx *gin.Context) {
	var newNotes []note
	if err := ctx.BindJSON(&newNotes); err != nil {
		return
	}

	var docs []interface{} = make([]interface{}, len(newNotes))
	for i, n := range newNotes {
		docs[i] = n
	}

	collection := client.Database("notes").Collection("notes")
	many, err := collection.InsertMany(context.TODO(), docs)
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.IndentedJSON(http.StatusCreated, many.InsertedIDs)
}

func getNoteById(ctx *gin.Context) {
	collection := client.Database("notes").Collection("notes")

	objID, err := bson.ObjectIDFromHex(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	filter := bson.M{"_id": objID}

	var result note
	err = collection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			ctx.JSON(http.StatusNotFound, gin.H{"message": "Note not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error finding note"})
		return
	}

	ctx.JSON(http.StatusOK, result)
}
