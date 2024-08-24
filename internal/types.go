package internal

import (
	"context"
	"reflect"
)

type Access interface {
	Get(ctx context.Context, sessionKey string, stateType reflect.Type) (any, error)
	Set(ctx context.Context, sessionKey string, state any) error
}
