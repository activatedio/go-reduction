package mux

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/activatedio/go-reduction"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/swaggest/openapi-go"
	"github.com/swaggest/openapi-go/openapi3"
	"net/http"
	"reflect"
)

var (
	contentTypeApplicationJSON = "application/json"
	cuOptionsJsonSuccess       = []openapi.ContentOption{openapi.WithContentType(contentTypeApplicationJSON), openapi.WithHTTPStatus(http.StatusOK)}
	cuOptionsJsonDefault       = []openapi.ContentOption{openapi.WithContentType(contentTypeApplicationJSON), func(cu *openapi.ContentUnit) {
		cu.IsDefault = true
		cu.Description = "Error"
	}}
)

type ReflectorBuilder func(rootPath string, reflector *openapi3.Reflector) error

type MountOptions struct {
	RootPath         string
	SwaggerRootPath  string
	ReflectorBuilder ReflectorBuilder
	// KeyFunc, when set, names whose state a request reads and writes (the
	// signed-in person, say), in place of the session cookie. An error or
	// an empty key answers 401 and runs nothing.
	KeyFunc func(r *http.Request) (string, error)
}

// withKey wraps h with opts.KeyFunc, when there is one.
func withKey(opts MountOptions, h http.HandlerFunc) http.HandlerFunc {
	if opts.KeyFunc == nil {
		return h
	}
	return func(w http.ResponseWriter, r *http.Request) {
		key, err := opts.KeyFunc(r)
		if err != nil || key == "" {
			writeError(w, http.StatusUnauthorized, &Error{Error: "unauthorized"})
			return
		}
		h(w, r.WithContext(reduction.WithKey(r.Context(), key)))
	}
}

func Mount(router *mux.Router, reduction reduction.Reduction, opts MountOptions) error {

	reflector := openapi3.NewReflector()

	reflector.SpecSchema().SetTitle("Reduction API")
	reflector.SpecSchema().SetVersion("v0.0.1")
	reflector.SpecSchema().SetDescription("Reduction API")

	for _, descriptor := range reduction.GetStateDescriptors() {
		statePath := opts.RootPath + descriptor.Path
		stateRoute := router.Path(statePath)
		stateRoute.Methods(http.MethodGet).HandlerFunc(withKey(opts, func(w http.ResponseWriter, r *http.Request) {

			log.Info().Msg("state method")

			ctx := r.Context()
			state, err := reduction.Get(ctx, descriptor.StateType)

			if err != nil {
				handleError(w, r, err)
				return
			}

			exported, err := descriptor.Exporter(ctx, state.State)

			if err != nil {
				// TODO - unit test this
				handleError(w, r, err)
				return
			}

			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			json.NewEncoder(w).Encode(exported)
		}))

		swaggerStatePath := opts.SwaggerRootPath + descriptor.Path
		if err := addStateOperation(swaggerStatePath, reflector, descriptor); err != nil {
			panic(err)
		}

		for _, a := range descriptor.Actions {
			actionPath := fmt.Sprintf("%s/%s", statePath, a.Path)
			router.Path(actionPath).Methods(http.MethodPost).HandlerFunc(withKey(opts, func(w http.ResponseWriter, r *http.Request) {
				log.Info().Msg("action method")

				action := reflect.New(a.ActionType).Interface()

				if !isEmpty(a.ActionType) {
					err := json.NewDecoder(r.Body).Decode(action)
					if err != nil {
						writeError(w, http.StatusBadRequest, &Error{Error: "the request body is not a valid " + a.Path})
						return
					}
				}

				ctx := r.Context()
				state, err := reduction.Set(ctx, descriptor.StateType, action)

				if err != nil {
					handleError(w, r, err)
					return
				}

				exported, err := descriptor.Exporter(ctx, state.State)

				if err != nil {
					// TODO - unit test this
					handleError(w, r, err)
					return
				}

				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				json.NewEncoder(w).Encode(exported)
			}))

			swaggerActionPath := fmt.Sprintf("%s/%s", swaggerStatePath, a.Path)
			if err := addActionOperation(swaggerActionPath, reflector, descriptor, a); err != nil {
				panic(err)
			}
		}
	}

	if opts.ReflectorBuilder != nil {
		err := opts.ReflectorBuilder(opts.SwaggerRootPath, reflector)
		if err != nil {
			return err
		}
	}

	swagger, err := reflector.Spec.MarshalJSON()
	if err != nil {
		return err
	}

	router.Path(fmt.Sprintf("%s/swagger.json", opts.RootPath)).Methods(http.MethodGet).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write(swagger)
	})

	return nil
}

func addStateOperation(path string, reflector *openapi3.Reflector, descriptor *reduction.StateDescriptor) error {

	oc, err := reflector.NewOperationContext(http.MethodGet, path)

	if err != nil {
		return err
	}

	oc.AddRespStructure(reflect.New(exportableType(descriptor.StateType)).Interface(), cuOptionsJsonSuccess...)
	oc.AddRespStructure(&Error{}, cuOptionsJsonDefault...)

	err = reflector.AddOperation(oc)

	if err != nil {
		return err
	}

	return nil
}

func addActionOperation(path string, reflector *openapi3.Reflector, stateDescriptor *reduction.StateDescriptor, actionDescriptor *reduction.ActionDescriptor) error {

	oc, err := reflector.NewOperationContext(http.MethodPost, path)

	if err != nil {
		return err
	}

	oc.AddRespStructure(reflect.New(exportableType(stateDescriptor.StateType)).Interface(), cuOptionsJsonSuccess...)
	oc.AddReqStructure(reflect.New(actionDescriptor.ActionType).Interface(), cuOptionsJsonSuccess...)
	oc.AddRespStructure(&Error{}, cuOptionsJsonDefault...)

	return reflector.AddOperation(oc)
}

var (
	emptyType = reflect.TypeFor[reduction.Empty]()
)

func isEmpty(t reflect.Type) bool {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		ft := f.Type
		if ft == emptyType && f.Anonymous {
			return true
		}
	}
	return false
}

// handleError answers a StatusError (a FieldError, a NotAllowedError, any
// error with HTTPStatus) with its status, and any other error with 500.
func handleError(w http.ResponseWriter, r *http.Request, err error) {
	var fe *reduction.FieldError
	if errors.As(err, &fe) {
		writeError(w, fe.HTTPStatus(), &Error{Error: fe.Message, Field: fe.Field})
		return
	}
	var se reduction.StatusError
	if errors.As(err, &se) {
		writeError(w, se.HTTPStatus(), &Error{Error: se.Error()})
		return
	}
	// Anything else is a fault on the server's side: its text (a database
	// error, an address, a path) never reaches the caller. The caller gets a
	// fixed sentence and a reference; the detail is logged once under it.
	ref := newRef()
	log.Error().Err(err).Str("ref", ref).Msg("reduction")
	writeError(w, http.StatusInternalServerError, &Error{Error: InternalErrorMessage + " (reference " + ref + ")", Ref: ref})
}

// InternalErrorMessage is what a caller is told about a server fault.
const InternalErrorMessage = "something went wrong on our side; try again"

// newRef is a short random reference for a logged fault.
func newRef() string {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "unknown"
	}
	return hex.EncodeToString(b)
}

func writeError(w http.ResponseWriter, status int, body *Error) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}

var (
	ExportMethodName = "Export"
)

func exportableType(typ reflect.Type) reflect.Type {
	if m, ok := reflect.PointerTo(typ).MethodByName(ExportMethodName); ok {
		return m.Func.Type().Out(0)
	} else {
		return typ
	}
}
