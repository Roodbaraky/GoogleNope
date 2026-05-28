package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func RequestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		ctx.Next()

		logger.InfoContext(ctx.Request.Context(), "http request",
			"method", ctx.Request.Method,
			"path", ctx.Request.URL.Path,
			"status", ctx.Writer.Status(),
			"latency_ms", time.Since(start).Milliseconds(),
			"client_ip", ctx.ClientIP(),
		)
	}
}

func RequestTimeout(timeout time.Duration) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if timeout <= 0 {
			ctx.Next()
			return
		}

		requestCtx, cancel := context.WithTimeout(ctx.Request.Context(), timeout)
		defer cancel()

		ctx.Request = ctx.Request.WithContext(requestCtx)
		ctx.Next()
	}
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func AbortWithError(ctx *gin.Context, status int, message string) {
	ctx.AbortWithStatusJSON(status, ErrorResponse{Error: message})
}

func MaxBodyBytes(limit int64) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if limit > 0 && ctx.Request.Body != nil {
			ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, limit)
		}

		ctx.Next()
	}
}

func CORS(allowedOrigins []string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	allowAll := false

	for _, origin := range allowedOrigins {
		if origin == "*" {
			allowAll = true
			continue
		}

		allowed[origin] = struct{}{}
	}

	return func(ctx *gin.Context) {
		origin := ctx.GetHeader("Origin")
		if origin != "" {
			if allowAll {
				ctx.Header("Access-Control-Allow-Origin", "*")
			} else if _, ok := allowed[origin]; ok {
				ctx.Header("Access-Control-Allow-Origin", origin)
				ctx.Header("Access-Control-Allow-Credentials", "true")
			}

			ctx.Header("Vary", "Origin")
			ctx.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
			ctx.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		}

		if ctx.Request.Method == http.MethodOptions {
			ctx.AbortWithStatus(http.StatusNoContent)
			return
		}

		ctx.Next()
	}
}
