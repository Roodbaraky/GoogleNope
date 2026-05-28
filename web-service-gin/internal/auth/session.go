package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
)

const (
	sessionCookieName = "googlenope_session"
	stateCookieName   = "googlenope_oauth_state"
)

var (
	ErrInvalidSession = errors.New("invalid session")
	ErrInvalidState   = errors.New("invalid oauth state")
)

type User struct {
	ID      string `json:"id"`
	Email   string `json:"email,omitempty"`
	Name    string `json:"name,omitempty"`
	Picture string `json:"picture,omitempty"`
}

type SessionManager struct {
	secret []byte
	secure bool
	now    func() time.Time
}

type sessionPayload struct {
	User      User  `json:"user"`
	ExpiresAt int64 `json:"expiresAt"`
}

type statePayload struct {
	State     string `json:"state"`
	ExpiresAt int64  `json:"expiresAt"`
}

func NewSessionManager(secret string, secure bool) (*SessionManager, error) {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return nil, errors.New("session secret is required")
	}

	return &SessionManager{
		secret: []byte(secret),
		secure: secure,
		now:    time.Now,
	}, nil
}

func (manager *SessionManager) SessionCookie(user User) (*http.Cookie, error) {
	payload := sessionPayload{
		User:      user,
		ExpiresAt: manager.now().Add(7 * 24 * time.Hour).Unix(),
	}

	value, err := manager.signJSON(payload)
	if err != nil {
		return nil, err
	}

	return manager.cookie(sessionCookieName, value, int((7 * 24 * time.Hour).Seconds())), nil
}

func (manager *SessionManager) ClearSessionCookie() *http.Cookie {
	cookie := manager.cookie(sessionCookieName, "", -1)
	cookie.Expires = time.Unix(0, 0)
	return cookie
}

func (manager *SessionManager) UserFromRequest(request *http.Request) (User, error) {
	cookie, err := request.Cookie(sessionCookieName)
	if err != nil {
		return User{}, ErrInvalidSession
	}

	var payload sessionPayload
	if err := manager.verifyJSON(cookie.Value, &payload); err != nil {
		return User{}, ErrInvalidSession
	}

	if payload.User.ID == "" || manager.now().Unix() > payload.ExpiresAt {
		return User{}, ErrInvalidSession
	}

	return payload.User, nil
}

func (manager *SessionManager) StateCookie(state string) (*http.Cookie, error) {
	payload := statePayload{
		State:     state,
		ExpiresAt: manager.now().Add(10 * time.Minute).Unix(),
	}

	value, err := manager.signJSON(payload)
	if err != nil {
		return nil, err
	}

	return manager.cookie(stateCookieName, value, int((10 * time.Minute).Seconds())), nil
}

func (manager *SessionManager) ClearStateCookie() *http.Cookie {
	cookie := manager.cookie(stateCookieName, "", -1)
	cookie.Expires = time.Unix(0, 0)
	return cookie
}

func (manager *SessionManager) VerifyState(request *http.Request, expected string) error {
	cookie, err := request.Cookie(stateCookieName)
	if err != nil {
		return ErrInvalidState
	}

	var payload statePayload
	if err := manager.verifyJSON(cookie.Value, &payload); err != nil {
		return ErrInvalidState
	}

	if payload.State == "" || payload.State != expected || manager.now().Unix() > payload.ExpiresAt {
		return ErrInvalidState
	}

	return nil
}

func (manager *SessionManager) cookie(name string, value string, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   manager.secure,
		SameSite: http.SameSiteLaxMode,
	}
}

func (manager *SessionManager) signJSON(payload any) (string, error) {
	encodedPayload, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	payloadPart := base64.RawURLEncoding.EncodeToString(encodedPayload)
	signature := manager.sign(payloadPart)
	return payloadPart + "." + signature, nil
}

func (manager *SessionManager) verifyJSON(value string, target any) error {
	payloadPart, signature, ok := strings.Cut(value, ".")
	if !ok || payloadPart == "" || signature == "" {
		return errors.New("malformed signed value")
	}

	if !hmac.Equal([]byte(signature), []byte(manager.sign(payloadPart))) {
		return errors.New("invalid signature")
	}

	payload, err := base64.RawURLEncoding.DecodeString(payloadPart)
	if err != nil {
		return err
	}

	return json.Unmarshal(payload, target)
}

func (manager *SessionManager) sign(value string) string {
	mac := hmac.New(sha256.New, manager.secret)
	mac.Write([]byte(value))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
