package internal

import (
	"context"
	"reflect"
)

type SetInternal func(ctx context.Context, state any, action any) (any, error)
type GetInternal func(ctx context.Context) (any, error)
type InitInternal func(ctx context.Context) (any, error)

func ToInitInternal(access Access, stateType reflect.Type, init any) InitInternal {
	return func(ctx context.Context) (any, error) {
		result := reflect.ValueOf(init).Call([]reflect.Value{reflect.ValueOf(ctx)})
		eInt := result[1].Interface()
		var err error
		if eInt != nil {
			err = eInt.(error)
		}
		return result[0].Interface(), err
	}
}

func ToGetInternal(access Access, stateType reflect.Type) GetInternal {
	return func(ctx context.Context) (any, error) {
		return access.Get(ctx, MustGetSessionID(ctx), stateType)
	}
}

func ToSetInternal(access Access, stateType reflect.Type, actionType reflect.Type, reducer any) SetInternal {
	return func(ctx context.Context, state any, action any) (any, error) {

		result := reflect.ValueOf(reducer).Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(state), reflect.ValueOf(action)})

		eInt := result[1].Interface()
		var err error
		if eInt != nil {
			err = eInt.(error)
		}

		if err != nil {
			return nil, err
		}

		resultState := result[0].Interface()

		err = access.Set(ctx, MustGetSessionID(ctx), stateType, resultState)

		if err != nil {
			return nil, err
		}

		return resultState, nil
	}
}
