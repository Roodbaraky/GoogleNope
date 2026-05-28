package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"

	"example/web-service-gin/internal/auth"
	"example/web-service-gin/internal/config"
	httpapi "example/web-service-gin/internal/http"
	"example/web-service-gin/internal/notes"
	mongostore "example/web-service-gin/internal/store/mongo"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("configuration error", "error", err)
		os.Exit(1)
	}

	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	rootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	connectCtx, cancelConnect := context.WithTimeout(rootCtx, cfg.MongoConnectTimeout)
	client, err := mongostore.Connect(connectCtx, cfg.MongoURI)
	cancelConnect()
	if err != nil {
		logger.Error("mongo connection error", "error", err)
		os.Exit(1)
	}

	defer func() {
		disconnectCtx, cancelDisconnect := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancelDisconnect()

		if err := client.Disconnect(disconnectCtx); err != nil {
			logger.Error("mongo disconnect error", "error", err)
		}
	}()

	database := client.Database(cfg.MongoDatabase)
	indexCtx, cancelIndexes := context.WithTimeout(rootCtx, cfg.MongoConnectTimeout)
	if err := mongostore.EnsureNoteIndexes(indexCtx, database, cfg.MongoCollection); err != nil {
		cancelIndexes()
		logger.Error("mongo index error", "error", err)
		os.Exit(1)
	}
	cancelIndexes()

	repository := mongostore.NewNotesRepository(database, cfg.MongoCollection, cfg.RequestTimeout)
	service := notes.NewService(repository)
	notesHandler := notes.NewHandler(service)
	sessionManager, err := auth.NewSessionManager(cfg.SessionSecret, cfg.Environment == "production")
	if err != nil {
		logger.Error("session configuration error", "error", err)
		os.Exit(1)
	}
	authHandler := auth.NewHandler(auth.OAuthConfig{
		ClientID:     cfg.OAuthClientID,
		ClientSecret: cfg.OAuthClientSecret,
		RedirectURL:  cfg.OAuthRedirectURL,
		AuthURL:      cfg.OAuthAuthURL,
		TokenURL:     cfg.OAuthTokenURL,
		UserInfoURL:  cfg.OAuthUserInfoURL,
	}, sessionManager)
	router := httpapi.NewRouter(httpapi.RouterDependencies{
		Config:       cfg,
		Logger:       logger,
		AuthHandler:  authHandler,
		AuthRequired: auth.RequireAuth(sessionManager),
		NotesHandler: notesHandler,
	})

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("starting api", "addr", cfg.HTTPAddr)
		errCh <- server.ListenAndServe()
	}()

	select {
	case <-rootCtx.Done():
		logger.Info("shutdown requested")
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancelShutdown()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown error", "error", err)
		os.Exit(1)
	}
}
