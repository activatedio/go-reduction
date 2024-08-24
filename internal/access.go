package internal

import (
	"context"
	"github.com/patrickmn/go-cache"
	"reflect"
)

type localAccess struct {
	cache cache.Cache
}

func (l *localAccess) Get(ctx context.Context, sessionKey string, stateType reflect.Type) (any, error) {
	//TODO implement me
	panic("implement me")
}

func (l *localAccess) Set(ctx context.Context, sessionKey string, stateType reflect.Type, state any) error {
	//TODO implement me
	panic("implement me")
}

// NewLocalAccess creates an access implementation which is appropriate for single node deployments
func NewLocalAccess() Access {
	return &localAccess{}
}
