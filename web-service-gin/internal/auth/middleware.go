package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const userContextKey = "auth.user"

func RequireAuth(manager *SessionManager) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user, err := manager.UserFromRequest(ctx.Request)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			return
		}

		SetUser(ctx, user)
		ctx.Next()
	}
}

func SetUser(ctx *gin.Context, user User) {
	ctx.Set(userContextKey, user)
}

func CurrentUser(ctx *gin.Context) (User, bool) {
	value, ok := ctx.Get(userContextKey)
	if !ok {
		return User{}, false
	}

	user, ok := value.(User)
	return user, ok && user.ID != ""
}

func CurrentUserID(ctx *gin.Context) (string, bool) {
	user, ok := CurrentUser(ctx)
	if !ok {
		return "", false
	}

	return user.ID, true
}
