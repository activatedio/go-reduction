package mux_test

import (
	"github.com/activatedio/reduction/mux"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

type stubNext struct {
	calls [][]any
}

func (s *stubNext) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	s.calls = append(s.calls, []any{writer, request})
}

func TestSessionMiddleware_Handle(t *testing.T) {

	type s struct {
		arrange func() *http.Request
		assert  func(r *http.Request)
	}

	hmacKey := "test-key"
	path := "/test-path"
	var sNext *stubNext
	var w *httptest.ResponseRecorder

	hmac := mux.NewHMAC(&mux.HMACConfig{Key: hmacKey})

	extractSessionID := func() string {
		cookieHeader := w.Header().Get("Set-Cookie")
		pattern, err := regexp.Compile("^_reduction_id=(.+); Path=/test-path; HttpOnly; SameSite=Lax$")
		check(err)
		matches := pattern.FindStringSubmatch(cookieHeader)
		assert.Len(t, matches, 2)
		return matches[1]
	}

	cases := map[string]s{
		"no session id": {
			arrange: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/", nil)
			},
			assert: func(r *http.Request) {
				assert.Len(t, sNext.calls, 1)
				sessionID := extractSessionID()
				ok, payload := hmac.ValidateAndExtract(sessionID)
				assert.True(t, ok)
				assert.Len(t, payload, 32)
				assert.NotEmpty(t, sessionID)
			},
		},
	}

	for k, v := range cases {
		t.Run(k, func(t *testing.T) {

			r := v.arrange()
			w = httptest.NewRecorder()
			sNext = &stubNext{}

			unit := mux.NewSessionMiddleware(&mux.SessionMiddlewareConfig{
				Secure:  false,
				Path:    path,
				HMACKey: hmacKey,
			}).Result

			unit.Handle(sNext).ServeHTTP(w, r)

			v.assert(r)
		})
	}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
