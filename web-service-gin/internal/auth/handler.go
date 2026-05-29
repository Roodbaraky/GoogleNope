package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

type OAuthConfig struct {
	ClientID           string
	ClientSecret       string
	RedirectURL        string
	AuthURL            string
	TokenURL           string
	UserInfoURL        string
	SuccessRedirectURL string
	AllowDevLogin      bool
}

type Handler struct {
	config  OAuthConfig
	manager *SessionManager
	client  *http.Client
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	IDToken     string `json:"id_token"`
}

type userInfoResponse struct {
	Subject string `json:"sub"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

func NewHandler(config OAuthConfig, manager *SessionManager) *Handler {
	return &Handler{
		config:  config,
		manager: manager,
		client:  http.DefaultClient,
	}
}

func (handler *Handler) RegisterRoutes(router gin.IRoutes) {
	router.GET("/login", handler.login)
	router.GET("/callback", handler.callback)
	router.POST("/logout", handler.logout)
	router.GET("/me", handler.me)
}

func (handler *Handler) login(ctx *gin.Context) {
	if err := handler.config.validate(); err != nil {
		if handler.config.AllowDevLogin {
			handler.devLogin(ctx)
			return
		}

		ctx.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "OAuth is not configured"})
		return
	}

	state, err := randomState()
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	cookie, err := handler.manager.StateCookie(state)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	http.SetCookie(ctx.Writer, cookie)
	ctx.Redirect(http.StatusFound, handler.authURL(state))
}

func (handler *Handler) callback(ctx *gin.Context) {
	if err := handler.config.validate(); err != nil {
		ctx.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "OAuth is not configured"})
		return
	}

	code := strings.TrimSpace(ctx.Query("code"))
	state := strings.TrimSpace(ctx.Query("state"))
	if code == "" || state == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid OAuth callback"})
		return
	}

	if err := handler.manager.VerifyState(ctx.Request, state); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid OAuth state"})
		return
	}

	token, err := handler.exchangeCode(ctx.Request.Context(), code)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadGateway, gin.H{"error": "OAuth token exchange failed"})
		return
	}

	user, err := handler.fetchUser(ctx.Request.Context(), token.AccessToken)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadGateway, gin.H{"error": "OAuth user lookup failed"})
		return
	}

	sessionCookie, err := handler.manager.SessionCookie(user)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	http.SetCookie(ctx.Writer, sessionCookie)
	http.SetCookie(ctx.Writer, handler.manager.ClearStateCookie())
	ctx.Redirect(http.StatusFound, handler.successRedirectURL())
}

func (handler *Handler) logout(ctx *gin.Context) {
	http.SetCookie(ctx.Writer, handler.manager.ClearSessionCookie())
	ctx.Status(http.StatusNoContent)
}

func (handler *Handler) me(ctx *gin.Context) {
	user, err := handler.manager.UserFromRequest(ctx.Request)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	SetUser(ctx, user)
	ctx.JSON(http.StatusOK, user)
}

func (handler *Handler) authURL(state string) string {
	values := url.Values{}
	values.Set("client_id", handler.config.ClientID)
	values.Set("redirect_uri", handler.config.RedirectURL)
	values.Set("response_type", "code")
	values.Set("scope", "openid email profile")
	values.Set("state", state)

	return handler.config.AuthURL + "?" + values.Encode()
}

func (handler *Handler) devLogin(ctx *gin.Context) {
	sessionCookie, err := handler.manager.SessionCookie(User{
		ID:    "dev:local-user",
		Email: "local@example.test",
		Name:  "Local Developer",
	})
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	http.SetCookie(ctx.Writer, sessionCookie)
	ctx.Redirect(http.StatusFound, handler.successRedirectURL())
}

func (handler *Handler) successRedirectURL() string {
	redirectURL := strings.TrimSpace(handler.config.SuccessRedirectURL)
	if redirectURL == "" {
		return "/"
	}

	return redirectURL
}

func (handler *Handler) exchangeCode(ctx context.Context, code string) (tokenResponse, error) {
	values := url.Values{}
	values.Set("client_id", handler.config.ClientID)
	values.Set("client_secret", handler.config.ClientSecret)
	values.Set("code", code)
	values.Set("grant_type", "authorization_code")
	values.Set("redirect_uri", handler.config.RedirectURL)

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, handler.config.TokenURL, strings.NewReader(values.Encode()))
	if err != nil {
		return tokenResponse{}, err
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := handler.client.Do(request)
	if err != nil {
		return tokenResponse{}, err
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return tokenResponse{}, fmt.Errorf("token endpoint returned %d", response.StatusCode)
	}

	var token tokenResponse
	if err := json.NewDecoder(response.Body).Decode(&token); err != nil {
		return tokenResponse{}, err
	}

	if strings.TrimSpace(token.AccessToken) == "" {
		return tokenResponse{}, errors.New("token response missing access token")
	}

	return token, nil
}

func (handler *Handler) fetchUser(ctx context.Context, accessToken string) (User, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, handler.config.UserInfoURL, nil)
	if err != nil {
		return User{}, err
	}

	request.Header.Set("Authorization", "Bearer "+accessToken)
	response, err := handler.client.Do(request)
	if err != nil {
		return User{}, err
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return User{}, fmt.Errorf("userinfo endpoint returned %d", response.StatusCode)
	}

	body := io.LimitReader(response.Body, 1<<20)
	var info userInfoResponse
	if err := json.NewDecoder(body).Decode(&info); err != nil {
		return User{}, err
	}

	if strings.TrimSpace(info.Subject) == "" {
		return User{}, errors.New("userinfo response missing subject")
	}

	return User{
		ID:      "oauth:" + info.Subject,
		Email:   info.Email,
		Name:    info.Name,
		Picture: info.Picture,
	}, nil
}

func (config OAuthConfig) validate() error {
	if strings.TrimSpace(config.ClientID) == "" ||
		strings.TrimSpace(config.ClientSecret) == "" ||
		strings.TrimSpace(config.RedirectURL) == "" ||
		strings.TrimSpace(config.AuthURL) == "" ||
		strings.TrimSpace(config.TokenURL) == "" ||
		strings.TrimSpace(config.UserInfoURL) == "" {
		return errors.New("oauth config is incomplete")
	}

	return nil
}

func randomState() (string, error) {
	var bytes [32]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(bytes[:]), nil
}
