package httpapi

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"

	"example/web-service-gin/internal/config"
	"example/web-service-gin/internal/notes"
)

func TestNewRouterHealthz(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := NewRouter(RouterDependencies{Config: testConfig()})
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}

	if response.Body.String() != `{"status":"ok"}` {
		t.Fatalf("unexpected health response: %s", response.Body.String())
	}
}

func TestNewRouterSetsCORSForAllowedOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := NewRouter(RouterDependencies{Config: testConfig()})
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	request.Header.Set("Origin", "http://localhost:4200")

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}

	if got := response.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:4200" {
		t.Fatalf("expected allowed origin header, got %q", got)
	}

	if got := response.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Fatalf("expected credentials header, got %q", got)
	}
}

func TestNewRouterReturnsJSONForMissingRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := NewRouter(RouterDependencies{Config: testConfig()})
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/missing", nil)

	router.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, response.Code)
	}

	if response.Body.String() != `{"error":"Not found"}` {
		t.Fatalf("unexpected error response: %s", response.Body.String())
	}
}

func TestNewRouterReturnsJSONForUnsupportedMethods(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := NewRouter(RouterDependencies{Config: testConfig()})
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/healthz", nil)

	router.ServeHTTP(response, request)

	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, response.Code)
	}

	if response.Body.String() != `{"error":"Method not allowed"}` {
		t.Fatalf("unexpected error response: %s", response.Body.String())
	}
}

func TestNewRouterAppliesMaxBodyBytes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := testConfig()
	cfg.MaxRequestBodyBytes = 8

	repository := &fakeRepository{}
	service := notes.NewService(repository)
	handler := notes.NewHandler(service)
	router := NewRouter(RouterDependencies{
		Config:       cfg,
		NotesHandler: handler,
	})
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/notes", bytes.NewBufferString(`[{"title":"too large"}]`))
	request.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}

	if len(repository.created) != 0 {
		t.Fatal("expected request body limit to prevent note creation")
	}
}

func testConfig() config.Config {
	return config.Config{
		Environment:         "test",
		HTTPAddr:            "localhost:0",
		MongoURI:            "mongodb://localhost:27017",
		MongoDatabase:       "notes",
		MongoCollection:     "notes",
		AllowedOrigins:      []string{"http://localhost:4200"},
		RequestTimeout:      time.Second,
		ShutdownTimeout:     time.Second,
		MongoConnectTimeout: time.Second,
		ReadHeaderTimeout:   time.Second,
		MaxRequestBodyBytes: 1 << 20,
	}
}

type fakeRepository struct {
	created []notes.Note
}

func (repository *fakeRepository) List(context.Context, notes.Pagination) ([]notes.Note, int64, error) {
	return []notes.Note{}, 0, nil
}

func (repository *fakeRepository) CreateMany(_ context.Context, newNotes []notes.Note) ([]notes.Note, error) {
	repository.created = append(repository.created, newNotes...)
	return newNotes, nil
}

func (repository *fakeRepository) FindByID(context.Context, bson.ObjectID) (notes.Note, error) {
	return notes.Note{}, notes.ErrNotFound
}
