package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHandlerLoginRedirectsToOAuthProvider(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := newAuthTestRouter(t)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/auth/login", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusFound {
		t.Fatalf("expected status %d, got %d", http.StatusFound, response.Code)
	}

	location := response.Header().Get("Location")
	for _, expected := range []string{
		"https://idp.example/auth?",
		"client_id=client-id",
		"redirect_uri=https%3A%2F%2Fapp.example%2Fapi%2Fauth%2Fcallback",
		"response_type=code",
		"scope=openid+email+profile",
		"state=",
	} {
		if !strings.Contains(location, expected) {
			t.Fatalf("expected redirect %q to contain %q", location, expected)
		}
	}

	if cookies := response.Result().Cookies(); len(cookies) == 0 {
		t.Fatal("expected OAuth state cookie")
	}
}

func TestHandlerLoginAllowsDevelopmentSessionWhenOAuthIsMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	manager := newTestSessionManager(t)
	router := gin.New()
	NewHandler(OAuthConfig{
		SuccessRedirectURL: "http://localhost:4200",
		AllowDevLogin:      true,
	}, manager).RegisterRoutes(router.Group("/api/auth"))

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/auth/login", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusFound {
		t.Fatalf("expected status %d, got %d: %s", http.StatusFound, response.Code, response.Body.String())
	}

	if response.Header().Get("Location") != "http://localhost:4200" {
		t.Fatalf("unexpected redirect: %q", response.Header().Get("Location"))
	}

	var sessionCookie *http.Cookie
	for _, cookie := range response.Result().Cookies() {
		if cookie.Name == sessionCookieName {
			sessionCookie = cookie
		}
	}
	if sessionCookie == nil {
		t.Fatal("expected session cookie")
	}

	request = httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	request.AddCookie(sessionCookie)
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected dev session to authenticate, got %d: %s", response.Code, response.Body.String())
	}
}

func TestHandlerLoginRejectsMissingOAuthOutsideDevLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	NewHandler(OAuthConfig{}, newTestSessionManager(t)).RegisterRoutes(router.Group("/api/auth"))

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/auth/login", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, response.Code)
	}
}

func TestHandlerMeReturnsCurrentSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	manager := newTestSessionManager(t)
	router := gin.New()
	NewHandler(testOAuthConfig(), manager).RegisterRoutes(router.Group("/api/auth"))

	cookie, err := manager.SessionCookie(User{ID: "oauth:123", Email: "user@example.com"})
	if err != nil {
		t.Fatalf("expected session cookie: %v", err)
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	request.AddCookie(cookie)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	if !strings.Contains(response.Body.String(), `"id":"oauth:123"`) {
		t.Fatalf("unexpected response: %s", response.Body.String())
	}
}

func TestHandlerMeRejectsMissingSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := newAuthTestRouter(t)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
}

func TestHandlerLogoutClearsSessionCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := newAuthTestRouter(t)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, response.Code)
	}

	cookies := response.Result().Cookies()
	if len(cookies) != 1 || cookies[0].MaxAge >= 0 {
		t.Fatalf("expected cleared session cookie, got %#v", cookies)
	}
}

func newAuthTestRouter(t *testing.T) *gin.Engine {
	t.Helper()

	router := gin.New()
	NewHandler(testOAuthConfig(), newTestSessionManager(t)).RegisterRoutes(router.Group("/api/auth"))
	return router
}

func testOAuthConfig() OAuthConfig {
	return OAuthConfig{
		ClientID:           "client-id",
		ClientSecret:       "client-secret",
		RedirectURL:        "https://app.example/api/auth/callback",
		AuthURL:            "https://idp.example/auth",
		TokenURL:           "https://idp.example/token",
		UserInfoURL:        "https://idp.example/userinfo",
		SuccessRedirectURL: "https://app.example",
	}
}
