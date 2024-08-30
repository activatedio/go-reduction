package fx

import (
	"github.com/activatedio/go-reduction"
	"github.com/activatedio/go-reduction/internal"
	"github.com/activatedio/go-reduction/mux"
	"go.uber.org/fx"
)

func Index() fx.Option {
	/*
		Potential options
		- omit config
	*/
	return fx.Module("activatedio.reduction",
		fx.Provide(internal.NewLocalAccessConfig, internal.NewLocalAccess, reduction.NewFactory, mux.NewSessionMiddlewareConfig, mux.NewSessionMiddleware),
	)
}
