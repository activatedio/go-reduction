package mux

import (
	"github.com/activatedio/reduction"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"net/http"
)

func Mount(router *mux.Router, rootPath string, reduction reduction.Reduction) error {

	route := router.PathPrefix(rootPath)

	for _, descriptor := range reduction.GetStateDescriptors() {
		stateRoute := route.Path(descriptor.Path)
		stateRoute.Methods(http.MethodGet).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			log.Info().Msg("state method")
			w.WriteHeader(http.StatusOK)
		})
		for _, a := range descriptor.Actions {
			stateRoute.Path(a.Path).Methods(http.MethodPost).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				log.Info().Msg("action method")
				w.WriteHeader(http.StatusOK)
			})
		}
	}

	return nil
}
