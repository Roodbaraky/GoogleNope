package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const (
	defaultEnvironment        = "development"
	defaultHTTPAddr           = "localhost:8080"
	defaultMongoDatabase      = "notes"
	defaultMongoCollection    = "notes"
	defaultAllowedOrigins     = "http://localhost:4200"
	defaultOAuthAuthURL       = "https://accounts.google.com/o/oauth2/v2/auth"
	defaultOAuthTokenURL      = "https://oauth2.googleapis.com/token"
	defaultOAuthUserInfoURL   = "https://openidconnect.googleapis.com/v1/userinfo"
	defaultRequestTimeout     = 5 * time.Second
	defaultShutdownTimeout    = 10 * time.Second
	defaultMongoConnectTime   = 10 * time.Second
	defaultReadHeaderTimeout  = 5 * time.Second
	defaultMaxRequestBodySize = int64(1 << 20)
)

type Config struct {
	Environment         string
	HTTPAddr            string
	MongoURI            string
	MongoDatabase       string
	MongoCollection     string
	AllowedOrigins      []string
	SessionSecret       string
	OAuthClientID       string
	OAuthClientSecret   string
	OAuthRedirectURL    string
	OAuthAuthURL        string
	OAuthTokenURL       string
	OAuthUserInfoURL    string
	RequestTimeout      time.Duration
	ShutdownTimeout     time.Duration
	MongoConnectTimeout time.Duration
	ReadHeaderTimeout   time.Duration
	MaxRequestBodyBytes int64
}

func Load() (Config, error) {
	_ = godotenv.Load(".env")

	requestTimeout, err := durationEnv("REQUEST_TIMEOUT", defaultRequestTimeout)
	if err != nil {
		return Config{}, err
	}

	shutdownTimeout, err := durationEnv("SHUTDOWN_TIMEOUT", defaultShutdownTimeout)
	if err != nil {
		return Config{}, err
	}

	mongoConnectTimeout, err := durationEnv("MONGO_CONNECT_TIMEOUT", defaultMongoConnectTime)
	if err != nil {
		return Config{}, err
	}

	readHeaderTimeout, err := durationEnv("READ_HEADER_TIMEOUT", defaultReadHeaderTimeout)
	if err != nil {
		return Config{}, err
	}

	maxRequestBodyBytes, err := int64Env("MAX_REQUEST_BODY_BYTES", defaultMaxRequestBodySize)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		Environment:         getEnv("APP_ENV", defaultEnvironment),
		HTTPAddr:            getEnv("HTTP_ADDR", defaultHTTPAddr),
		MongoURI:            strings.TrimSpace(os.Getenv("MONGODB_URI")),
		MongoDatabase:       getEnv("MONGODB_DATABASE", defaultMongoDatabase),
		MongoCollection:     getEnv("MONGODB_COLLECTION", defaultMongoCollection),
		AllowedOrigins:      splitCSV(getEnv("CORS_ALLOWED_ORIGINS", defaultAllowedOrigins)),
		SessionSecret:       strings.TrimSpace(os.Getenv("SESSION_SECRET")),
		OAuthClientID:       strings.TrimSpace(os.Getenv("OAUTH_CLIENT_ID")),
		OAuthClientSecret:   strings.TrimSpace(os.Getenv("OAUTH_CLIENT_SECRET")),
		OAuthRedirectURL:    strings.TrimSpace(os.Getenv("OAUTH_REDIRECT_URL")),
		OAuthAuthURL:        getEnv("OAUTH_AUTH_URL", defaultOAuthAuthURL),
		OAuthTokenURL:       getEnv("OAUTH_TOKEN_URL", defaultOAuthTokenURL),
		OAuthUserInfoURL:    getEnv("OAUTH_USERINFO_URL", defaultOAuthUserInfoURL),
		RequestTimeout:      requestTimeout,
		ShutdownTimeout:     shutdownTimeout,
		MongoConnectTimeout: mongoConnectTimeout,
		ReadHeaderTimeout:   readHeaderTimeout,
		MaxRequestBodyBytes: maxRequestBodyBytes,
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (cfg Config) Validate() error {
	if strings.TrimSpace(cfg.MongoURI) == "" {
		return errors.New("MONGODB_URI is required")
	}

	if strings.TrimSpace(cfg.HTTPAddr) == "" {
		return errors.New("HTTP_ADDR is required")
	}

	if strings.TrimSpace(cfg.MongoDatabase) == "" {
		return errors.New("MONGODB_DATABASE is required")
	}

	if strings.TrimSpace(cfg.MongoCollection) == "" {
		return errors.New("MONGODB_COLLECTION is required")
	}

	if strings.TrimSpace(cfg.SessionSecret) == "" {
		return errors.New("SESSION_SECRET is required")
	}

	if cfg.Environment == "production" {
		if strings.TrimSpace(cfg.OAuthClientID) == "" {
			return errors.New("OAUTH_CLIENT_ID is required in production")
		}

		if strings.TrimSpace(cfg.OAuthClientSecret) == "" {
			return errors.New("OAUTH_CLIENT_SECRET is required in production")
		}

		if strings.TrimSpace(cfg.OAuthRedirectURL) == "" {
			return errors.New("OAUTH_REDIRECT_URL is required in production")
		}
	}

	if strings.TrimSpace(cfg.OAuthAuthURL) == "" {
		return errors.New("OAUTH_AUTH_URL is required")
	}

	if strings.TrimSpace(cfg.OAuthTokenURL) == "" {
		return errors.New("OAUTH_TOKEN_URL is required")
	}

	if strings.TrimSpace(cfg.OAuthUserInfoURL) == "" {
		return errors.New("OAUTH_USERINFO_URL is required")
	}

	if cfg.RequestTimeout <= 0 {
		return fmt.Errorf("REQUEST_TIMEOUT must be positive")
	}

	if cfg.ShutdownTimeout <= 0 {
		return fmt.Errorf("SHUTDOWN_TIMEOUT must be positive")
	}

	if cfg.MongoConnectTimeout <= 0 {
		return fmt.Errorf("MONGO_CONNECT_TIMEOUT must be positive")
	}

	if cfg.ReadHeaderTimeout <= 0 {
		return fmt.Errorf("READ_HEADER_TIMEOUT must be positive")
	}

	if cfg.MaxRequestBodyBytes <= 0 {
		return fmt.Errorf("MAX_REQUEST_BODY_BYTES must be positive")
	}

	return nil
}

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	return value
}

func durationEnv(key string, fallback time.Duration) (time.Duration, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid duration: %w", key, err)
	}

	return parsed, nil
}

func int64Env(key string, fallback int64) (int64, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", key, err)
	}

	return parsed, nil
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item != "" {
			result = append(result, item)
		}
	}

	return result
}
