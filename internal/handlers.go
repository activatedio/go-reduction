package internal

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"reflect"
)

type SetInternal func(ctx context.Context, state any, action any) (any, error)
type GetInternal func(ctx context.Context) (any, error)
type InitInternal func(ctx context.Context) (any, error)
type RefreshInternal func(ctx context.Context, state any) (any, error)
type ExportInternal func(ctx context.Context, state any) (any, error)

const (
	OperationInit    = "init"
	OperationGet     = "get"
	OperationSet     = "set"
	OperationRefresh = "refresh"
	OperationExport  = "export"
)

type logEvent struct {
	operation  string
	stateType  reflect.Type
	actionType reflect.Type
	stateIn    any
	stateOut   any
	action     any
	err        error
}

type logOption func(e *logEvent)

func withOperation(operation string) logOption {
	return func(e *logEvent) {
		e.operation = operation
	}
}

func withStateType(t reflect.Type) logOption {
	return func(e *logEvent) {
		e.stateType = t
	}
}

func withActionType(t reflect.Type) logOption {
	return func(e *logEvent) {
		e.actionType = t
	}
}

func withStateIn(state any) logOption {
	return func(e *logEvent) {
		if state != nil {
			e.stateIn = state
		}
	}
}

func withStateOut(state any) logOption {
	return func(e *logEvent) {
		if state != nil {
			e.stateIn = state
		}
	}
}

func withAction(action any) logOption {
	return func(e *logEvent) {
		e.action = action
	}
}

func withError(err error) logOption {
	return func(e *logEvent) {
		e.err = err
	}
}

func logInternal(ctx context.Context, opts ...logOption) {

	evt := &logEvent{}

	for _, o := range opts {
		o(evt)
	}

	var ze *zerolog.Event

	if evt.err != nil {
		ze = log.Error().Err(evt.err)
	} else {
		ze = log.Debug()
	}

	ze.Str("session_id", MustGetSessionID(ctx))

	if evt.stateIn != nil {
		ze = ze.Str("state_type", reflect.TypeOf(evt.stateIn).Elem().Name()).Interface("state_in", evt.stateIn)
		if evt.stateOut != nil {
			ze = ze.Interface("state_out", evt.stateOut)
		}
	} else if evt.stateOut != nil {
		ze = ze.Str("state_type", reflect.TypeOf(evt.stateOut).Elem().Name()).Interface("state_out", evt.stateOut)
	}

	if evt.stateIn == nil && evt.stateOut == nil {
		ze = ze.Str("state_type", evt.stateType.Name())
	}

	if evt.action != nil {
		ze = ze.Str("action_type", reflect.TypeOf(evt.action).Elem().Name()).Interface("action", evt.action)
	} else if evt.actionType != nil {
		ze = ze.Str("action_type", evt.actionType.Name())
	}

	msg := func() string {
		if evt.err != nil {
			return "error"
		} else {
			return "success"
		}
	}

	if evt.operation != "" {
		ze = ze.Str("operation", evt.operation)
	}
	ze.Msg(msg())
}

type logBuilder struct {
	opts []logOption
}

func lb() *logBuilder {
	return &logBuilder{}
}

func (l *logBuilder) opt(opts ...logOption) *logBuilder {
	l.opts = append(l.opts, opts...)
	return l
}

func (l *logBuilder) log(ctx context.Context) {
	logInternal(ctx, l.opts...)
}

func ToInitInternal(access Access, stateType reflect.Type, init any) InitInternal {
	return func(ctx context.Context) (any, error) {
		_lb := lb().opt(withOperation(OperationInit), withStateType(stateType))
		result := reflect.ValueOf(init).Call([]reflect.Value{reflect.ValueOf(ctx)})
		eInt := result[1].Interface()
		var err error
		if eInt != nil {
			err = eInt.(error)
			_lb.opt(withError(err)).log(ctx)
			return nil, err
		}
		res := result[0].Interface()
		_lb.opt(withStateOut(res)).log(ctx)
		return result[0].Interface(), nil
	}
}

func ToGetInternal(access Access, stateType reflect.Type) GetInternal {
	return func(ctx context.Context) (any, error) {
		res, err := access.Get(ctx, MustGetSessionID(ctx), stateType)
		lb().opt(withOperation(OperationGet), withStateType(stateType), withStateOut(res), withError(err)).log(ctx)
		return res, err
	}
}

func ToRefreshInternal(access Access, stateType reflect.Type, refresher any) RefreshInternal {
	return func(ctx context.Context, state any) (any, error) {

		_lb := lb().opt(withOperation(OperationRefresh), withStateIn(state))

		result := reflect.ValueOf(refresher).Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(state)})

		return doSet(ctx, access, result, _lb)
	}
}

func ToSetInternal(access Access, stateType reflect.Type, actionType reflect.Type, reducer any) SetInternal {
	return func(ctx context.Context, state any, action any) (any, error) {

		_lb := lb().opt(withOperation(OperationSet), withStateIn(state), withAction(action))

		result := reflect.ValueOf(reducer).Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(state), reflect.ValueOf(action)})

		return doSet(ctx, access, result, _lb)
	}
}

func doSet(ctx context.Context, access Access, result []reflect.Value, _lb *logBuilder) (any, error) {

	eInt := result[1].Interface()
	var err error
	if eInt != nil {
		err = eInt.(error)
		_lb.opt(withError(err)).log(ctx)
		return nil, err
	}

	resultState := result[0].Interface()

	err = access.Set(ctx, MustGetSessionID(ctx), resultState)

	if err != nil {
		_lb.opt(withError(err)).log(ctx)
		return nil, err
	}

	_lb.opt(withStateOut(resultState)).log(ctx)
	return resultState, nil
}

var (
	ExportMethodName = "Export"
)

func ToExportInternal(stateType reflect.Type) ExportInternal {

	_, ok := reflect.PointerTo(stateType).MethodByName(ExportMethodName)

	if ok {

		return func(ctx context.Context, state any) (any, error) {

			_lb := lb().opt(withOperation(OperationExport), withStateIn(state))

			result := reflect.ValueOf(state).MethodByName(ExportMethodName).Call([]reflect.Value{reflect.ValueOf(ctx)})

			eInt := result[1].Interface()
			var err error
			if eInt != nil {
				err = eInt.(error)
				_lb.opt(withError(err)).log(ctx)
				return nil, err
			}
			res := result[0].Interface()
			_lb.opt(withStateOut(res)).log(ctx)
			return res, nil
		}

	} else {
		return func(ctx context.Context, state any) (any, error) {
			return state, nil
		}
	}
}
