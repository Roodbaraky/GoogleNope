package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
)

func main() {
	router := gin.Default()
	router.GET("/notes", getNotes)
	router.POST("/notes", postNotes)
	router.GET("/notes/:id", getNoteById)

	router.Run("localhost:8080")
}

type note struct {
	ID      uuid.UUID `json:"id"`
	Title   string    `json:"title"`
	Content string    `json:"content"`
}

var notes = []note{
	{
		ID:      uuid.New(),
		Title:   "Note 1",
		Content: "Content 1",
	},
	{
		ID:      uuid.New(),
		Title:   "Note 2",
		Content: "Content 2",
	},
	{
		ID:      uuid.New(),
		Title:   "Note 3",
		Content: "Content 3",
	},
}

func getNotes(context *gin.Context) {
	context.JSON(http.StatusOK, notes)
}

func postNotes(context *gin.Context) {
	var newNotes []note
	if err := context.BindJSON(&newNotes); err != nil {
		return
	}

	notes = append(notes, newNotes...)
	context.IndentedJSON(http.StatusCreated, newNotes)
}

func getNoteById(context *gin.Context) {
	for _, note := range notes {
		if note.ID.String() == context.Param("id") {
			context.IndentedJSON(http.StatusOK, note)
			return
		}
	}
	context.IndentedJSON(http.StatusNotFound, gin.H{"message": "Note not found"})
	return
}
