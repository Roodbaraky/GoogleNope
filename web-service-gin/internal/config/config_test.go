package config

import (
	"strings"
	"testing"
	"time"
)

func TestLoadRequiresMongoURI(t *testing.T) {
	clearConfigEnv(t)

	_, err := Load()
	if err == nil {
		t.Fatal("expected missing Mongo URI to fail")
	}

	if !strings.Contains(err.Error(), "MONGODB_URI") {
		t.Fatalf("expected MONGODB_URI error, got %q", err.Error())
	}
}

func TestLoadRequiresSessionSecret(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("MONGODB_URI", "mongodb://localhost:27017")

	_, err := Load()
	if err == nil {
		t.Fatal("expected missing session secret to fail")
	}

	if !strings.Contains(err.Error(), "SESSION_SECRET") {
		t.Fatalf("expected SESSION_SECRET error, got %q", err.Error())
	}
}

func TestLoadUsesDefaults(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("MONGODB_URI", "mongodb://localhost:27017")
	t.Setenv("SESSION_SECRET", "test-session-secret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected config to load: %v", err)
	}

	if cfg.Environment != defaultEnvironment {
		t.Fatalf("expected default environment %q, got %q", defaultEnvironment, cfg.Environment)
	}

	if cfg.HTTPAddr != defaultHTTPAddr {
		t.Fatalf("expected default HTTP addr %q, got %q", defaultHTTPAddr, cfg.HTTPAddr)
	}

	if cfg.FrontendURL != defaultFrontendURL {
		t.Fatalf("expected default frontend URL %q, got %q", defaultFrontendURL, cfg.FrontendURL)
	}

	if cfg.MongoDatabase != defaultMongoDatabase {
		t.Fatalf("expected default Mongo database %q, got %q", defaultMongoDatabase, cfg.MongoDatabase)
	}

	if cfg.MongoCollection != defaultMongoCollection {
		t.Fatalf("expected default Mongo collection %q, got %q", defaultMongoCollection, cfg.MongoCollection)
	}

	if len(cfg.AllowedOrigins) != 1 || cfg.AllowedOrigins[0] != defaultAllowedOrigins {
		t.Fatalf("expected default allowed origin %q, got %#v", defaultAllowedOrigins, cfg.AllowedOrigins)
	}

	if cfg.RequestTimeout != defaultRequestTimeout {
		t.Fatalf("expected default request timeout %s, got %s", defaultRequestTimeout, cfg.RequestTimeout)
	}

	if cfg.MaxRequestBodyBytes != defaultMaxRequestBodySize {
		t.Fatalf("expected default max body size %d, got %d", defaultMaxRequestBodySize, cfg.MaxRequestBodyBytes)
	}

	if cfg.SessionSecret != "test-session-secret" {
		t.Fatalf("expected session secret to load")
	}

	if cfg.OAuthAuthURL != defaultOAuthAuthURL || cfg.OAuthTokenURL != defaultOAuthTokenURL || cfg.OAuthUserInfoURL != defaultOAuthUserInfoURL {
		t.Fatalf("expected default OAuth provider URLs")
	}
}

func TestLoadUsesOverrides(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("APP_ENV", "production")
	t.Setenv("HTTP_ADDR", ":9090")
	t.Setenv("FRONTEND_URL", "https://app.example")
	t.Setenv("MONGODB_URI", "mongodb://db.example:27017")
	t.Setenv("MONGODB_DATABASE", "google_nope")
	t.Setenv("MONGODB_COLLECTION", "user_notes")
	t.Setenv("SESSION_SECRET", "test-session-secret")
	t.Setenv("OAUTH_CLIENT_ID", "client-id")
	t.Setenv("OAUTH_CLIENT_SECRET", "client-secret")
	t.Setenv("OAUTH_REDIRECT_URL", "https://app.example/api/auth/callback")
	t.Setenv("OAUTH_AUTH_URL", "https://idp.example/auth")
	t.Setenv("OAUTH_TOKEN_URL", "https://idp.example/token")
	t.Setenv("OAUTH_USERINFO_URL", "https://idp.example/userinfo")
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:4200, https://app.example")
	t.Setenv("REQUEST_TIMEOUT", "3s")
	t.Setenv("SHUTDOWN_TIMEOUT", "4s")
	t.Setenv("MONGO_CONNECT_TIMEOUT", "5s")
	t.Setenv("READ_HEADER_TIMEOUT", "6s")
	t.Setenv("MAX_REQUEST_BODY_BYTES", "2048")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected config to load: %v", err)
	}

	if cfg.Environment != "production" {
		t.Fatalf("expected production environment, got %q", cfg.Environment)
	}

	if cfg.HTTPAddr != ":9090" {
		t.Fatalf("expected HTTP addr override, got %q", cfg.HTTPAddr)
	}

	if cfg.FrontendURL != "https://app.example" {
		t.Fatalf("expected frontend URL override, got %q", cfg.FrontendURL)
	}

	if cfg.MongoURI != "mongodb://db.example:27017" {
		t.Fatalf("expected Mongo URI override, got %q", cfg.MongoURI)
	}

	if cfg.MongoDatabase != "google_nope" {
		t.Fatalf("expected Mongo database override, got %q", cfg.MongoDatabase)
	}

	if cfg.MongoCollection != "user_notes" {
		t.Fatalf("expected Mongo collection override, got %q", cfg.MongoCollection)
	}

	expectedOrigins := []string{"http://localhost:4200", "https://app.example"}
	if len(cfg.AllowedOrigins) != len(expectedOrigins) {
		t.Fatalf("expected %d origins, got %#v", len(expectedOrigins), cfg.AllowedOrigins)
	}
	for i, origin := range expectedOrigins {
		if cfg.AllowedOrigins[i] != origin {
			t.Fatalf("expected origin %q at index %d, got %q", origin, i, cfg.AllowedOrigins[i])
		}
	}

	if cfg.RequestTimeout != 3*time.Second {
		t.Fatalf("expected request timeout override, got %s", cfg.RequestTimeout)
	}

	if cfg.ShutdownTimeout != 4*time.Second {
		t.Fatalf("expected shutdown timeout override, got %s", cfg.ShutdownTimeout)
	}

	if cfg.MongoConnectTimeout != 5*time.Second {
		t.Fatalf("expected Mongo connect timeout override, got %s", cfg.MongoConnectTimeout)
	}

	if cfg.ReadHeaderTimeout != 6*time.Second {
		t.Fatalf("expected read header timeout override, got %s", cfg.ReadHeaderTimeout)
	}

	if cfg.MaxRequestBodyBytes != 2048 {
		t.Fatalf("expected max body bytes override, got %d", cfg.MaxRequestBodyBytes)
	}

	if cfg.OAuthClientID != "client-id" || cfg.OAuthClientSecret != "client-secret" {
		t.Fatalf("expected OAuth client settings to load")
	}

	if cfg.OAuthRedirectURL != "https://app.example/api/auth/callback" {
		t.Fatalf("expected OAuth redirect URL override, got %q", cfg.OAuthRedirectURL)
	}

	if cfg.OAuthAuthURL != "https://idp.example/auth" ||
		cfg.OAuthTokenURL != "https://idp.example/token" ||
		cfg.OAuthUserInfoURL != "https://idp.example/userinfo" {
		t.Fatalf("expected OAuth URL overrides")
	}
}

func TestLoadFailsOnMalformedDuration(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("MONGODB_URI", "mongodb://localhost:27017")
	t.Setenv("SESSION_SECRET", "test-session-secret")
	t.Setenv("REQUEST_TIMEOUT", "not-a-duration")

	_, err := Load()
	if err == nil {
		t.Fatal("expected malformed duration to fail")
	}

	if !strings.Contains(err.Error(), "REQUEST_TIMEOUT") {
		t.Fatalf("expected REQUEST_TIMEOUT error, got %q", err.Error())
	}
}

func clearConfigEnv(t *testing.T) {
	t.Helper()

	for _, key := range []string{
		"APP_ENV",
		"HTTP_ADDR",
		"FRONTEND_URL",
		"MONGODB_URI",
		"MONGODB_DATABASE",
		"MONGODB_COLLECTION",
		"SESSION_SECRET",
		"OAUTH_CLIENT_ID",
		"OAUTH_CLIENT_SECRET",
		"OAUTH_REDIRECT_URL",
		"OAUTH_AUTH_URL",
		"OAUTH_TOKEN_URL",
		"OAUTH_USERINFO_URL",
		"CORS_ALLOWED_ORIGINS",
		"REQUEST_TIMEOUT",
		"SHUTDOWN_TIMEOUT",
		"MONGO_CONNECT_TIMEOUT",
		"READ_HEADER_TIMEOUT",
		"MAX_REQUEST_BODY_BYTES",
	} {
		t.Setenv(key, "")
	}
}
