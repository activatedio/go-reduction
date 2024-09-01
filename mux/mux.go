package mux

import (
	"encoding/json"
	"fmt"
	"github.com/activatedio/go-reduction"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"net/http"
	"reflect"
)

func Mount(router *mux.Router, rootPath string, reduction reduction.Reduction) error {

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
		}
	}

	return nil
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
