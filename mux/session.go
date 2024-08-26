package mux

import (
	"github.com/activatedio/reduction/config"
	"github.com/activatedio/reduction/internal"
	"github.com/activatedio/reduction/util"
	"go.uber.org/fx"
	"math/rand"
	"net/http"
)

const (
	SessionIDCookieName = "_reduction_id"
)

type sessionMiddleware struct {
	hmac         HMAC
	cookieSecure bool
	cookiePath   string
}

func (s *sessionMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		signedSessionID := s.getSignedSessionIDFromCookie(r)
		var sessionID string

		if signedSessionID != "" {
			ok, tmp := s.hmac.ValidateAndExtract(signedSessionID)
			if ok {
				sessionID = tmp
			}
		}

		if sessionID == "" {
			sessionID = s.generateAndSetSessionIDCookie(w)
		}

		next.ServeHTTP(w, r.WithContext(internal.WithSessionID(r.Context(), sessionID)))
	})
}

type SessionMiddlewareResult struct {
	fx.Out
	Result Middleware `name:"middleware"`
}

type SessionMiddlewareConfig struct {
	HMACKey string
	Secure  bool
	Path    string
}

func NewSessionMiddlewareConfig() *SessionMiddlewareConfig {
	return &SessionMiddlewareConfig{
		HMACKey: util.MustGetEnv(config.ReductionKeySessionHMACKey),
		Secure:  util.GetEnvBool(config.ReductionKeySessionSecure, true),
		Path:    util.GetEnv(config.ReductionKeySessionCookiePath, "/"),
	}
}

func NewSessionMiddleware(config *SessionMiddlewareConfig) SessionMiddlewareResult {
	return SessionMiddlewareResult{
		Result: &sessionMiddleware{
			hmac:         NewHMAC(&HMACConfig{Key: config.HMACKey}),
			cookieSecure: config.Secure,
			cookiePath:   config.Path,
		},
	}
}

func (s *sessionMiddleware) getSignedSessionIDFromCookie(r *http.Request) string {

	c, err := r.Cookie(SessionIDCookieName)

	if err != nil {
		return ""
	} else {
		return c.Value
	}
}

func (s *sessionMiddleware) generateAndSetSessionIDCookie(w http.ResponseWriter) string {

	value := RandSeq(32)

	cookie := &http.Cookie{
		Name:     SessionIDCookieName,
		Value:    s.hmac.Sign(value),
		Path:     s.cookiePath,
		HttpOnly: true,
		Secure:   s.cookieSecure,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, cookie)
	return value
}

var letters = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

func RandSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
