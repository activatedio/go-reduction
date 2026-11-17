package internal

import (
	"context"
	"reflect"
)

// Access keeps a state per session key and state type. Get answers nil, nil
// when there is none. It is exported as reduction.Store.
type Access interface {
	Get(ctx context.Context, sessionKey string, stateType reflect.Type) (any, error)
	Set(ctx context.Context, sessionKey string, state any) error
}

// TxAccess is an Access that can run a read, a reducer and the write of its
// result as one unit. It is exported as reduction.TxStore.
type TxAccess interface {
	Access
	InTx(ctx context.Context, fn func(ctx context.Context) error) error
}
