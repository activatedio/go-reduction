package mux_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/activatedio/go-reduction"
	rmux "github.com/activatedio/go-reduction/mux"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Steps struct {
	Step string `json:"step"`
	Name string `json:"name,omitempty"`
}

type SetName struct {
	Name string `json:"name"`
}

type Forbidden struct {
	reduction.Empty
}

type Broken struct {
	reduction.Empty
}

func mountSteps(t *testing.T) http.Handler {
	r := reduction.New(reduction.NewMemoryStore(time.Minute))
	r.Builder().State(reflect.TypeFor[Steps]()).
		Init(func(ctx context.Context) (*Steps, error) {
			return &Steps{Step: "name"}, nil
		}).
		Action(reflect.TypeFor[SetName](), func(ctx context.Context, s *Steps, a *SetName) (*Steps, error) {
			if strings.TrimSpace(a.Name) == "" {
				return nil, reduction.Invalid("name", "a name is required")
			}
			return &Steps{Step: "done", Name: a.Name}, nil
		}).
		Action(reflect.TypeFor[Forbidden](), func(ctx context.Context, s *Steps, a *Forbidden) (*Steps, error) {
			return nil, reduction.NotAllowed("not yours")
		}).
		Action(reflect.TypeFor[Broken](), func(ctx context.Context, s *Steps, a *Broken) (*Steps, error) {
			return nil, errors.New("broken")
		})
	router := mux.NewRouter()
	require.NoError(t, rmux.Mount(router, r, rmux.MountOptions{
		RootPath: "/api",
		KeyFunc: func(r *http.Request) (string, error) {
			who := r.Header.Get("X-Who")
			if who == "error" {
				return "", errors.New("bad bearer")
			}
			return who, nil
		},
	}))
	return router
}

func call(h http.Handler, method, path, who, body string) (int, map[string]any) {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if who != "" {
		req.Header.Set("X-Who", who)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	out := map[string]any{}
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	return rec.Code, out
}

func TestMount_KeyFunc(t *testing.T) {
	h := mountSteps(t)

	code, body := call(h, http.MethodGet, "/api/steps", "", "")
	assert.Equal(t, http.StatusUnauthorized, code)
	assert.Equal(t, "unauthorized", body["error"])

	code, _ = call(h, http.MethodGet, "/api/steps", "error", "")
	assert.Equal(t, http.StatusUnauthorized, code)

	code, _ = call(h, http.MethodPost, "/api/steps/set_name", "", `{"name":"x"}`)
	assert.Equal(t, http.StatusUnauthorized, code)

	code, body = call(h, http.MethodPost, "/api/steps/set_name", "alice", `{"name":"Acme"}`)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, map[string]any{"step": "done", "name": "Acme"}, body)

	// No cookie: the key alone finds the state again, and only for its owner.
	_, body = call(h, http.MethodGet, "/api/steps", "alice", "")
	assert.Equal(t, "done", body["step"])
	_, body = call(h, http.MethodGet, "/api/steps", "bob", "")
	assert.Equal(t, "name", body["step"])
}

func TestMount_ErrorStatuses(t *testing.T) {
	h := mountSteps(t)

	code, body := call(h, http.MethodPost, "/api/steps/set_name", "alice", `{"name":" "}`)
	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, map[string]any{"error": "a name is required", "field": "name"}, body)

	code, body = call(h, http.MethodPost, "/api/steps/set_name", "alice", `{not json`)
	assert.Equal(t, http.StatusBadRequest, code)
	assert.NotEmpty(t, body["error"])

	code, body = call(h, http.MethodPost, "/api/steps/forbidden", "alice", "")
	assert.Equal(t, http.StatusForbidden, code)
	assert.Equal(t, map[string]any{"error": "not yours"}, body)

	code, body = call(h, http.MethodPost, "/api/steps/broken", "alice", "")
	assert.Equal(t, http.StatusInternalServerError, code)
	assert.Equal(t, map[string]any{"error": "broken"}, body)

	// The failed actions left the state as it was.
	_, body = call(h, http.MethodGet, "/api/steps", "alice", "")
	assert.Equal(t, "name", body["step"])
}

func TestMount_SwaggerHasErrorField(t *testing.T) {
	h := mountSteps(t)
	req := httptest.NewRequest(http.MethodGet, "/api/swagger.json", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"field"`)
}
