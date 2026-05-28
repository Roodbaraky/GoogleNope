package httpapi

import (
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"example/web-service-gin/internal/auth"
	"example/web-service-gin/internal/config"
	"example/web-service-gin/internal/notes"
)

type RouterDependencies struct {
	Config       config.Config
	Logger       *slog.Logger
	AuthHandler  *auth.Handler
	AuthRequired gin.HandlerFunc
	NotesHandler *notes.Handler
}

func NewRouter(deps RouterDependencies) *gin.Engine {
	logger := deps.Logger
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}

	router := gin.New()
	router.HandleMethodNotAllowed = true
	router.Use(gin.Recovery())
	router.Use(RequestLogger(logger))
	router.Use(MaxBodyBytes(deps.Config.MaxRequestBodyBytes))
	router.Use(RequestTimeout(deps.Config.RequestTimeout))
	router.Use(TrustedOrigins(deps.Config.AllowedOrigins))
	router.Use(CORS(deps.Config.AllowedOrigins))

	router.NoRoute(func(ctx *gin.Context) {
		AbortWithError(ctx, http.StatusNotFound, "Not found")
	})

	router.NoMethod(func(ctx *gin.Context) {
		AbortWithError(ctx, http.StatusMethodNotAllowed, "Method not allowed")
	})

	router.GET("/healthz", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	if deps.AuthHandler != nil {
		authRoutes := router.Group("/api/auth")
		deps.AuthHandler.RegisterRoutes(authRoutes)
	}

	if deps.NotesHandler != nil {
		notesRoutes := router.Group("/api/notes")
		if deps.AuthRequired != nil {
			notesRoutes.Use(deps.AuthRequired)
		}

		deps.NotesHandler.RegisterRoutes(notesRoutes)
	}

	return router
}
