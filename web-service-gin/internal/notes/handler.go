package notes

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
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
}

func (handler *Handler) getNotes(ctx *gin.Context) {
	page, limit, ok := parsePagination(ctx)
	if !ok {
		abortWithError(ctx, http.StatusBadRequest, "Invalid pagination parameters")
		return
	}

	pageInfo, err := handler.service.List(ctx.Request.Context(), page, limit)
	if err != nil {
		internalServerError(ctx)
		return
	}

	ctx.JSON(http.StatusOK, pageInfo)
}

func (handler *Handler) postNotes(ctx *gin.Context) {
	var newNotes []Note
	if err := ctx.ShouldBindJSON(&newNotes); err != nil {
		abortWithError(ctx, http.StatusBadRequest, "Invalid JSON request body")
		return
	}

	storedNotes, err := handler.service.CreateMany(ctx.Request.Context(), newNotes)
	if err != nil {
		if errors.Is(err, ErrInvalidInput) {
			abortWithError(ctx, http.StatusBadRequest, err.Error())
			return
		}

		internalServerError(ctx)
		return
	}

	ctx.JSON(http.StatusCreated, storedNotes)
}

func (handler *Handler) getNoteByID(ctx *gin.Context) {
	note, err := handler.service.FindByID(ctx.Request.Context(), ctx.Param("id"))
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
