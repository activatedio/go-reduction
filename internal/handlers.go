package internal

import (
	"context"
	"reflect"
)

type SetInternal func(ctx context.Context, state any, action any) (any, error)
type GetInternal func(ctx context.Context) (any, error)
type InitInternal func(ctx context.Context) (any, error)
type RefreshInternal func(ctx context.Context, state any) (any, error)
type ExportInternal func(ctx context.Context, state any) (any, error)

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

func ToRefreshInternal(access Access, stateType reflect.Type, refresher any) RefreshInternal {
	return func(ctx context.Context, state any) (any, error) {

		result := reflect.ValueOf(refresher).Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(state)})

		return doSet(ctx, access, result)
	}
}

func ToSetInternal(access Access, stateType reflect.Type, actionType reflect.Type, reducer any) SetInternal {
	return func(ctx context.Context, state any, action any) (any, error) {

		result := reflect.ValueOf(reducer).Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(state), reflect.ValueOf(action)})

		return doSet(ctx, access, result)
	}
}

func doSet(ctx context.Context, access Access, result []reflect.Value) (any, error) {

	eInt := result[1].Interface()
	var err error
	if eInt != nil {
		err = eInt.(error)
	}

	if err != nil {
		return nil, err
	}

	resultState := result[0].Interface()

	err = access.Set(ctx, MustGetSessionID(ctx), resultState)

	if err != nil {
		return nil, err
	}

	return resultState, nil
}

var (
	ExportMethodName = "Export"
)

func ToExportInternal(stateType reflect.Type) ExportInternal {

	_, ok := reflect.PointerTo(stateType).MethodByName(ExportMethodName)

	if ok {

		return func(ctx context.Context, state any) (any, error) {
			result := reflect.ValueOf(state).MethodByName(ExportMethodName).Call([]reflect.Value{reflect.ValueOf(ctx)})

			eInt := result[1].Interface()
			var err error
			if eInt != nil {
				err = eInt.(error)
			}
			return result[0].Interface(), err
		}

	} else {
		return func(ctx context.Context, state any) (any, error) {
			return state, nil
		}
	}
}
