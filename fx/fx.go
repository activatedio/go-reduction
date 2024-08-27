package fx

import (
	"github.com/activatedio/reduction"
	"github.com/activatedio/reduction/internal"
	"github.com/activatedio/reduction/mux"
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
