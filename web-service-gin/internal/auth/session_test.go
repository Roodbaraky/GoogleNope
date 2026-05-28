package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestSessionManagerRoundTripsSignedCookie(t *testing.T) {
	manager := newTestSessionManager(t)
	user := User{ID: "oauth:123", Email: "user@example.com"}

	cookie, err := manager.SessionCookie(user)
	if err != nil {
		t.Fatalf("expected session cookie: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.AddCookie(cookie)

	got, err := manager.UserFromRequest(request)
	if err != nil {
		t.Fatalf("expected valid session: %v", err)
	}

	if got.ID != user.ID || got.Email != user.Email {
		t.Fatalf("unexpected user: %#v", got)
	}
}

func TestSessionManagerRejectsTamperedCookie(t *testing.T) {
	manager := newTestSessionManager(t)
	cookie, err := manager.SessionCookie(User{ID: "oauth:123"})
	if err != nil {
		t.Fatalf("expected session cookie: %v", err)
	}
	cookie.Value += "tampered"

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.AddCookie(cookie)

	if _, err := manager.UserFromRequest(request); err == nil {
		t.Fatal("expected tampered session to fail")
	}
}

func TestRequireAuthSetsCurrentUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	manager := newTestSessionManager(t)
	cookie, err := manager.SessionCookie(User{ID: "oauth:123"})
	if err != nil {
		t.Fatalf("expected session cookie: %v", err)
	}

	router := gin.New()
	router.Use(RequireAuth(manager))
	router.GET("/protected", func(ctx *gin.Context) {
		userID, ok := CurrentUserID(ctx)
		if !ok {
			t.Fatal("expected current user")
		}
		ctx.JSON(http.StatusOK, gin.H{"id": userID})
	})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	request.AddCookie(cookie)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}

	if response.Body.String() != `{"id":"oauth:123"}` {
		t.Fatalf("unexpected response: %s", response.Body.String())
	}
}

func TestRequireAuthRejectsMissingSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	manager := newTestSessionManager(t)

	router := gin.New()
	router.Use(RequireAuth(manager))
	router.GET("/protected", func(ctx *gin.Context) {
		t.Fatal("handler should not run")
	})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
}

func TestSessionManagerVerifiesOAuthState(t *testing.T) {
	manager := newTestSessionManager(t)
	cookie, err := manager.StateCookie("state-123")
	if err != nil {
		t.Fatalf("expected state cookie: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/callback", nil)
	request.AddCookie(cookie)

	if err := manager.VerifyState(request, "state-123"); err != nil {
		t.Fatalf("expected valid state: %v", err)
	}

	if err := manager.VerifyState(request, "other-state"); err == nil {
		t.Fatal("expected mismatched state to fail")
	}
}

func newTestSessionManager(t *testing.T) *SessionManager {
	t.Helper()

	manager, err := NewSessionManager("test-session-secret", false)
	if err != nil {
		t.Fatalf("expected session manager: %v", err)
	}
	manager.now = func() time.Time {
		return time.Date(2026, 5, 28, 10, 0, 0, 0, time.UTC)
	}

	return manager
}
