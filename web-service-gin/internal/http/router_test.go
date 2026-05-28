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

	"example/web-service-gin/internal/auth"
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

func TestNewRouterRejectsUnsafeDisallowedOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := NewRouter(RouterDependencies{Config: testConfig()})
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	request.Header.Set("Origin", "https://evil.example")

	router.ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, response.Code)
	}

	if response.Body.String() != `{"error":"Origin not allowed"}` {
		t.Fatalf("unexpected error response: %s", response.Body.String())
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
		AuthRequired: testAuthRequired(),
		NotesHandler: handler,
	})
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/notes", bytes.NewBufferString(`{"title":"too large"}`))
	request.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}

	if repository.created {
		t.Fatal("expected request body limit to prevent note creation")
	}
}

func TestNewRouterProtectsNoteRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repository := &fakeRepository{}
	service := notes.NewService(repository)
	handler := notes.NewHandler(service)
	manager, err := auth.NewSessionManager("test-session-secret", false)
	if err != nil {
		t.Fatalf("expected session manager: %v", err)
	}
	router := NewRouter(RouterDependencies{
		Config:       testConfig(),
		AuthRequired: auth.RequireAuth(manager),
		NotesHandler: handler,
	})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/notes", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
}

func testConfig() config.Config {
	return config.Config{
		Environment:         "test",
		HTTPAddr:            "localhost:0",
		MongoURI:            "mongodb://localhost:27017",
		MongoDatabase:       "notes",
		MongoCollection:     "notes",
		SessionSecret:       "test-session-secret",
		OAuthAuthURL:        "https://idp.example/auth",
		OAuthTokenURL:       "https://idp.example/token",
		OAuthUserInfoURL:    "https://idp.example/userinfo",
		AllowedOrigins:      []string{"http://localhost:4200"},
		RequestTimeout:      time.Second,
		ShutdownTimeout:     time.Second,
		MongoConnectTimeout: time.Second,
		ReadHeaderTimeout:   time.Second,
		MaxRequestBodyBytes: 1 << 20,
	}
}

func testAuthRequired() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		auth.SetUser(ctx, auth.User{ID: "oauth:test-user"})
		ctx.Next()
	}
}

type fakeRepository struct {
	created bool
}

func (repository *fakeRepository) List(context.Context, string, notes.Pagination) ([]notes.Note, int64, error) {
	return []notes.Note{}, 0, nil
}

func (repository *fakeRepository) Create(_ context.Context, newNote notes.Note) (notes.Note, error) {
	repository.created = true
	return newNote, nil
}

func (repository *fakeRepository) FindByID(context.Context, string, bson.ObjectID) (notes.Note, error) {
	return notes.Note{}, notes.ErrNotFound
}

func (repository *fakeRepository) Update(context.Context, string, bson.ObjectID, notes.NoteUpdate) (notes.Note, error) {
	return notes.Note{}, notes.ErrNotFound
}

func (repository *fakeRepository) Delete(context.Context, string, bson.ObjectID) error {
	return notes.ErrNotFound
}
