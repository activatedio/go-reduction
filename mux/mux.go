package mux

import (
	"encoding/json"
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

type MountOptions struct {
	RootPath         string
	SwaggerRootPath  string
	ReflectorBuilder func(rootPath string, reflector *openapi3.Reflector) error
}

func Mount(router *mux.Router, reduction reduction.Reduction, opts MountOptions) error {

	reflector := openapi3.NewReflector()

	reflector.SpecSchema().SetTitle("Reduction API")
	reflector.SpecSchema().SetVersion("v0.0.1")
	reflector.SpecSchema().SetDescription("Reduction API")

	for _, descriptor := range reduction.GetStateDescriptors() {
		statePath := opts.RootPath + descriptor.Path
		stateRoute := router.Path(statePath)
		stateRoute.Methods(http.MethodGet).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

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
		})

		swaggerStatePath := opts.SwaggerRootPath + descriptor.Path
		if err := addStateOperation(swaggerStatePath, reflector, descriptor); err != nil {
			panic(err)
		}

		for _, a := range descriptor.Actions {
			actionPath := fmt.Sprintf("%s/%s", statePath, a.Path)
			router.Path(actionPath).Methods(http.MethodPost).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				log.Info().Msg("action method")

				action := reflect.New(a.ActionType).Interface()

				if !isEmpty(a.ActionType) {
					err := json.NewDecoder(r.Body).Decode(action)
					if err != nil {
						handleError(w, r, err)
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
			})

			swaggerActionPath := fmt.Sprintf("%s/%s", swaggerStatePath, a.Path)
			if err := addActionOperation(swaggerActionPath, reflector, descriptor, a); err != nil {
				panic(err)
			}
		}
	}

	if opts.ReflectorBuilder != nil {
		err := opts.ReflectorBuilder(opts.RootPath, reflector)
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

func handleError(w http.ResponseWriter, r *http.Request, err error) {
	log.Error().Err(err)
	w.Header().Set("Content-Type", "application/json;")
	w.WriteHeader(http.StatusInternalServerError)
	// TODO - let's make this better
	json.NewEncoder(w).Encode(&Error{Error: err.Error()})
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
