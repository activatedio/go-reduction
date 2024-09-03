package mux

import (
	"encoding/json"
	"fmt"
	"github.com/activatedio/go-reduction"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/swaggest/openapi-go/openapi3"
	"net/http"
	"reflect"
)

func Mount(router *mux.Router, rootPath string, reduction reduction.Reduction) error {

	reflector := openapi3.NewReflector()

	reflector.SpecSchema().SetTitle("Reduction API")
	reflector.SpecSchema().SetVersion("v0.0.1")
	reflector.SpecSchema().SetDescription("Reduction API")

	for _, descriptor := range reduction.GetStateDescriptors() {
		statePath := rootPath + descriptor.Path
		stateRoute := router.Path(statePath)
		stateRoute.Methods(http.MethodGet).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			log.Info().Msg("state method")

			ctx := r.Context()
			state, err := reduction.Get(ctx, descriptor.StateType)

			if err != nil {
				handleError(w, r, err)
				return
			}

			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			json.NewEncoder(w).Encode(state.State)
		})

		if err := addStateOperation(statePath, reflector, descriptor); err != nil {
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

				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				json.NewEncoder(w).Encode(state.State)
			})

			if err := addActionOperation(actionPath, reflector, descriptor, a); err != nil {
				panic(err)
			}
		}
	}

	swagger, err := reflector.Spec.MarshalJSON()
	if err != nil {
		return err
	}

	router.Path(fmt.Sprintf("%s/swagger.json", rootPath)).Methods(http.MethodGet).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	oc.AddRespStructure(reflect.New(descriptor.StateType).Interface())

	return reflector.AddOperation(oc)
}

func addActionOperation(path string, reflector *openapi3.Reflector, stateDescriptor *reduction.StateDescriptor, actionDescriptor *reduction.ActionDescriptor) error {

	oc, err := reflector.NewOperationContext(http.MethodPost, path)

	if err != nil {
		return err
	}

	oc.AddRespStructure(reflect.New(stateDescriptor.StateType).Interface())
	oc.AddReqStructure(reflect.New(actionDescriptor.ActionType).Interface())

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
	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
}
