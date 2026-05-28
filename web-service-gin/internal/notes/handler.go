package notes

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"example/web-service-gin/internal/auth"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (handler *Handler) RegisterRoutes(router gin.IRoutes) {
	router.GET("", handler.getNotes)
	router.POST("", handler.postNotes)
	router.GET("/:id", handler.getNoteByID)
	router.PATCH("/:id", handler.patchNote)
	router.DELETE("/:id", handler.deleteNote)
}

func (handler *Handler) getNotes(ctx *gin.Context) {
	page, limit, ok := parsePagination(ctx)
	if !ok {
		abortWithError(ctx, http.StatusBadRequest, "Invalid pagination parameters")
		return
	}

	userID, ok := currentUserID(ctx)
	if !ok {
		unauthorized(ctx)
		return
	}

	pageInfo, err := handler.service.List(ctx.Request.Context(), userID, page, limit)
	if err != nil {
		internalServerError(ctx)
		return
	}

	ctx.JSON(http.StatusOK, pageInfo)
}

func (handler *Handler) postNotes(ctx *gin.Context) {
	var input CreateNoteInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		abortWithError(ctx, http.StatusBadRequest, "Invalid JSON request body")
		return
	}

	userID, ok := currentUserID(ctx)
	if !ok {
		unauthorized(ctx)
		return
	}

	storedNote, err := handler.service.Create(ctx.Request.Context(), userID, input)
	if err != nil {
		if errors.Is(err, ErrInvalidInput) {
			abortWithError(ctx, http.StatusBadRequest, "Invalid note")
			return
		}

		internalServerError(ctx)
		return
	}

	ctx.JSON(http.StatusCreated, storedNote)
}

func (handler *Handler) getNoteByID(ctx *gin.Context) {
	userID, ok := currentUserID(ctx)
	if !ok {
		unauthorized(ctx)
		return
	}

	note, err := handler.service.FindByID(ctx.Request.Context(), userID, ctx.Param("id"))
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidID):
			abortWithError(ctx, http.StatusBadRequest, "Invalid ID format")
		case errors.Is(err, ErrNotFound):
			abortWithError(ctx, http.StatusNotFound, "Note not found")
		default:
			internalServerError(ctx)
		}

		return
	}

	ctx.JSON(http.StatusOK, note)
}

func (handler *Handler) patchNote(ctx *gin.Context) {
	var input UpdateNoteInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		abortWithError(ctx, http.StatusBadRequest, "Invalid JSON request body")
		return
	}

	userID, ok := currentUserID(ctx)
	if !ok {
		unauthorized(ctx)
		return
	}

	note, err := handler.service.Update(ctx.Request.Context(), userID, ctx.Param("id"), input)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidID):
			abortWithError(ctx, http.StatusBadRequest, "Invalid ID format")
		case errors.Is(err, ErrInvalidInput):
			abortWithError(ctx, http.StatusBadRequest, "Invalid note")
		case errors.Is(err, ErrNotFound):
			abortWithError(ctx, http.StatusNotFound, "Note not found")
		default:
			internalServerError(ctx)
		}

		return
	}

	ctx.JSON(http.StatusOK, note)
}

func (handler *Handler) deleteNote(ctx *gin.Context) {
	userID, ok := currentUserID(ctx)
	if !ok {
		unauthorized(ctx)
		return
	}

	err := handler.service.Delete(ctx.Request.Context(), userID, ctx.Param("id"))
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidID):
			abortWithError(ctx, http.StatusBadRequest, "Invalid ID format")
		case errors.Is(err, ErrNotFound):
			abortWithError(ctx, http.StatusNotFound, "Note not found")
		default:
			internalServerError(ctx)
		}

		return
	}

	ctx.Status(http.StatusNoContent)
}

func parsePagination(ctx *gin.Context) (int, int, bool) {
	limitStr := ctx.DefaultQuery("limit", "10")
	pageStr := ctx.DefaultQuery("page", "1")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		return 0, 0, false
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		return 0, 0, false
	}

	return page, limit, true
}

type errorResponse struct {
	Error string `json:"error"`
}

func abortWithError(ctx *gin.Context, status int, message string) {
	ctx.AbortWithStatusJSON(status, errorResponse{Error: message})
}

func internalServerError(ctx *gin.Context) {
	abortWithError(ctx, http.StatusInternalServerError, "Internal server error")
}

func unauthorized(ctx *gin.Context) {
	abortWithError(ctx, http.StatusUnauthorized, "Authentication required")
}

func currentUserID(ctx *gin.Context) (string, bool) {
	return auth.CurrentUserID(ctx)
}
