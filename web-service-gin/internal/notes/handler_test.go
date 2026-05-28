package notes

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"

	"example/web-service-gin/internal/auth"
)

const testUserID = "oauth:test-user"

func TestHandlerGetNotesReturnsPaginatedNotes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	noteID := bson.NewObjectID()
	repository := &handlerFakeRepository{
		listNotes: []Note{
			{
				ID:        noteID,
				UserID:    testUserID,
				Title:     "Pinned",
				Content:   "Content",
				Pinned:    true,
				CreatedAt: time.Date(2026, 5, 28, 9, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2026, 5, 28, 9, 30, 0, 0, time.UTC),
			},
		},
		listTotal: 10,
	}
	router := newTestNotesRouter(repository)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/notes?page=2&limit=3", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var page Page
	if err := json.Unmarshal(response.Body.Bytes(), &page); err != nil {
		t.Fatalf("expected page response: %v", err)
	}

	if repository.listUserID != testUserID {
		t.Fatalf("expected development user id, got %q", repository.listUserID)
	}

	if repository.listPagination.Page != 2 || repository.listPagination.Limit != 3 {
		t.Fatalf("unexpected pagination: %#v", repository.listPagination)
	}

	if page.Total != 10 || page.Page != 2 || page.Limit != 3 || page.Pages != 4 {
		t.Fatalf("unexpected page metadata: %#v", page)
	}

	if len(page.Notes) != 1 || page.Notes[0].ID != noteID || !page.Notes[0].Pinned {
		t.Fatalf("unexpected notes response: %#v", page.Notes)
	}
}

func TestHandlerGetNotesRejectsInvalidPagination(t *testing.T) {
	router := newTestNotesRouter(&handlerFakeRepository{})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/notes?page=0&limit=10", nil)
	router.ServeHTTP(response, request)

	assertJSONError(t, response, http.StatusBadRequest, "Invalid pagination parameters")
}

func TestHandlerPostNotesCreatesNote(t *testing.T) {
	repository := &handlerFakeRepository{}
	router := newTestNotesRouter(repository)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/notes", strings.NewReader(`{"title":"  Title  ","content":"Body","pinned":true}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}

	if repository.created.UserID != testUserID {
		t.Fatalf("expected created note to be user scoped, got %q", repository.created.UserID)
	}

	if repository.created.Title != "Title" || repository.created.Content != "Body" || !repository.created.Pinned {
		t.Fatalf("unexpected created note: %#v", repository.created)
	}

	var note Note
	if err := json.Unmarshal(response.Body.Bytes(), &note); err != nil {
		t.Fatalf("expected note response: %v", err)
	}

	if note.ID == (bson.ObjectID{}) || note.Title != "Title" || note.UserID != "" {
		t.Fatalf("unexpected response note: %#v", note)
	}
}

func TestHandlerPostNotesRejectsInvalidNote(t *testing.T) {
	repository := &handlerFakeRepository{}
	router := newTestNotesRouter(repository)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/notes", strings.NewReader(`{"title":"   ","content":""}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(response, request)

	assertJSONError(t, response, http.StatusBadRequest, "Invalid note")
	if repository.createCalled {
		t.Fatal("expected invalid note to be rejected before repository call")
	}
}

func TestHandlerGetNoteByIDReturnsNote(t *testing.T) {
	noteID := bson.NewObjectID()
	repository := &handlerFakeRepository{
		findNote: Note{
			ID:      noteID,
			UserID:  testUserID,
			Title:   "Found",
			Content: "Body",
		},
	}
	router := newTestNotesRouter(repository)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/notes/"+noteID.Hex(), nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	if repository.findUserID != testUserID || repository.findID != noteID {
		t.Fatalf("unexpected find scope: user=%q id=%s", repository.findUserID, repository.findID.Hex())
	}

	var note Note
	if err := json.Unmarshal(response.Body.Bytes(), &note); err != nil {
		t.Fatalf("expected note response: %v", err)
	}

	if note.ID != noteID || note.Title != "Found" {
		t.Fatalf("unexpected note response: %#v", note)
	}
}

func TestHandlerGetNoteByIDHandlesErrors(t *testing.T) {
	t.Run("invalid id", func(t *testing.T) {
		router := newTestNotesRouter(&handlerFakeRepository{})

		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/api/notes/not-an-id", nil)
		router.ServeHTTP(response, request)

		assertJSONError(t, response, http.StatusBadRequest, "Invalid ID format")
	})

	t.Run("not found", func(t *testing.T) {
		repository := &handlerFakeRepository{findErr: ErrNotFound}
		router := newTestNotesRouter(repository)

		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/api/notes/"+bson.NewObjectID().Hex(), nil)
		router.ServeHTTP(response, request)

		assertJSONError(t, response, http.StatusNotFound, "Note not found")
	})
}

func TestHandlerPatchNoteUpdatesNote(t *testing.T) {
	noteID := bson.NewObjectID()
	repository := &handlerFakeRepository{}
	router := newTestNotesRouter(repository)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPatch, "/api/notes/"+noteID.Hex(), strings.NewReader(`{"title":" Updated ","pinned":true}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	if repository.updateUserID != testUserID || repository.updateID != noteID {
		t.Fatalf("unexpected update scope: user=%q id=%s", repository.updateUserID, repository.updateID.Hex())
	}

	if repository.update.Title == nil || *repository.update.Title != "Updated" {
		t.Fatalf("expected trimmed title update, got %#v", repository.update.Title)
	}

	if repository.update.Pinned == nil || !*repository.update.Pinned {
		t.Fatal("expected pinned update")
	}
}

func TestHandlerPatchNoteHandlesErrors(t *testing.T) {
	t.Run("empty patch", func(t *testing.T) {
		router := newTestNotesRouter(&handlerFakeRepository{})

		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPatch, "/api/notes/"+bson.NewObjectID().Hex(), strings.NewReader(`{}`))
		request.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(response, request)

		assertJSONError(t, response, http.StatusBadRequest, "Invalid note")
	})

	t.Run("not found", func(t *testing.T) {
		repository := &handlerFakeRepository{updateErr: ErrNotFound}
		router := newTestNotesRouter(repository)

		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPatch, "/api/notes/"+bson.NewObjectID().Hex(), strings.NewReader(`{"content":"Body"}`))
		request.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(response, request)

		assertJSONError(t, response, http.StatusNotFound, "Note not found")
	})
}

func TestHandlerDeleteNoteDeletesNote(t *testing.T) {
	noteID := bson.NewObjectID()
	repository := &handlerFakeRepository{}
	router := newTestNotesRouter(repository)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodDelete, "/api/notes/"+noteID.Hex(), nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d: %s", http.StatusNoContent, response.Code, response.Body.String())
	}

	if repository.deleteUserID != testUserID || repository.deleteID != noteID {
		t.Fatalf("unexpected delete scope: user=%q id=%s", repository.deleteUserID, repository.deleteID.Hex())
	}
}

func TestHandlerDeleteNoteHandlesErrors(t *testing.T) {
	t.Run("invalid id", func(t *testing.T) {
		router := newTestNotesRouter(&handlerFakeRepository{})

		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodDelete, "/api/notes/not-an-id", nil)
		router.ServeHTTP(response, request)

		assertJSONError(t, response, http.StatusBadRequest, "Invalid ID format")
	})

	t.Run("not found", func(t *testing.T) {
		repository := &handlerFakeRepository{deleteErr: ErrNotFound}
		router := newTestNotesRouter(repository)

		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodDelete, "/api/notes/"+bson.NewObjectID().Hex(), nil)
		router.ServeHTTP(response, request)

		assertJSONError(t, response, http.StatusNotFound, "Note not found")
	})
}

func newTestNotesRouter(repository Repository) *gin.Engine {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(func(ctx *gin.Context) {
		auth.SetUser(ctx, auth.User{ID: testUserID, Email: "test@example.com"})
		ctx.Next()
	})
	handler := NewHandler(NewService(repository))
	handler.RegisterRoutes(router.Group("/api/notes"))
	return router
}

func assertJSONError(t *testing.T, response *httptest.ResponseRecorder, status int, message string) {
	t.Helper()

	if response.Code != status {
		t.Fatalf("expected status %d, got %d: %s", status, response.Code, response.Body.String())
	}

	expected := `{"error":"` + message + `"}`
	if response.Body.String() != expected {
		t.Fatalf("expected error %s, got %s", expected, response.Body.String())
	}
}

type handlerFakeRepository struct {
	listNotes      []Note
	listTotal      int64
	listUserID     string
	listPagination Pagination

	createCalled bool
	created      Note

	findNote   Note
	findErr    error
	findUserID string
	findID     bson.ObjectID

	update       NoteUpdate
	updateErr    error
	updateUserID string
	updateID     bson.ObjectID

	deleteErr    error
	deleteUserID string
	deleteID     bson.ObjectID
}

func (repository *handlerFakeRepository) List(_ context.Context, userID string, pagination Pagination) ([]Note, int64, error) {
	repository.listUserID = userID
	repository.listPagination = pagination
	return repository.listNotes, repository.listTotal, nil
}

func (repository *handlerFakeRepository) Create(_ context.Context, newNote Note) (Note, error) {
	repository.createCalled = true
	repository.created = newNote
	newNote.ID = bson.NewObjectID()
	return newNote, nil
}

func (repository *handlerFakeRepository) FindByID(_ context.Context, userID string, id bson.ObjectID) (Note, error) {
	repository.findUserID = userID
	repository.findID = id
	if repository.findErr != nil {
		return Note{}, repository.findErr
	}

	return repository.findNote, nil
}

func (repository *handlerFakeRepository) Update(_ context.Context, userID string, id bson.ObjectID, update NoteUpdate) (Note, error) {
	repository.updateUserID = userID
	repository.updateID = id
	repository.update = update
	if repository.updateErr != nil {
		return Note{}, repository.updateErr
	}

	return Note{
		ID:        id,
		UserID:    userID,
		Title:     stringValue(update.Title),
		Content:   stringValue(update.Content),
		Pinned:    boolValue(update.Pinned),
		UpdatedAt: update.UpdatedAt,
	}, nil
}

func (repository *handlerFakeRepository) Delete(_ context.Context, userID string, id bson.ObjectID) error {
	repository.deleteUserID = userID
	repository.deleteID = id
	return repository.deleteErr
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}

func boolValue(value *bool) bool {
	return value != nil && *value
}
